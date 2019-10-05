package server

import (
	"fmt"

	pb "git.corp.angel-salon.com/gc/proto"
)

type validator struct {
	collectionByName map[string]*pb.Collection
	specTypes        map[string]map[string]pb.Type
	fieldType        map[string]map[string]pb.Type // collectionName : (fieldName : type)
}

func validate(pipeline *pb.Pipeline, specs []*pb.TableMetadata, checkResult bool) error {
	v := &validator{}
	v.fieldType = make(map[string]map[string]pb.Type)

	v.specTypes = make(map[string]map[string]pb.Type)
	for _, spec := range specs {
		specType := make(map[string]pb.Type)
		for _, field := range spec.Fields {
			specType[field.Name] = field.Type
		}
		v.specTypes[spec.Name] = specType
	}

	// 表名唯一
	collectionByName := make(map[string]*pb.Collection)
	for _, c := range pipeline.Collections {
		if _, existed := collectionByName[c.Name]; existed {
			return fmt.Errorf("有一个或多个表叫做 %v", c.Name)
		}
		collectionByName[c.Name] = c
	}
	v.collectionByName = collectionByName

	// 不支持取样
	for _, c := range pipeline.Collections {
		switch c.Body.(type) {
		case *pb.Collection_Sample:
			return fmt.Errorf("不支持取样")
		}
	}

	seq, err := v.validateTopology(pipeline)
	if err != nil {
		return err
	}
	pipeline.Collections = seq

	for _, c := range pipeline.Collections {
		t, err := v.validateType(c)
		if err != nil {
			return err
		}
		v.fieldType[c.Name] = t
	}

	if checkResult {
		lastCollection := pipeline.Collections[len(pipeline.Collections)-1]
		for field, _type := range v.fieldType[lastCollection.Name] {
			if specType, ok := v.specTypes[pipeline.ResultMetaName]; ok {
				if columnType, ok := specType[field]; ok {
					if _type != columnType {
						return fmt.Errorf("结果表的列：%s 类型不符", field)
					}
				} else {
					return fmt.Errorf("结果表多余列:%s", field)
				}
			} else {
				return fmt.Errorf("结果表的表模板：%s 不存在", pipeline.ResultMetaName)
			}
		}
		if len(v.fieldType[lastCollection.Name]) != len(v.specTypes[pipeline.ResultMetaName]) {
			names := ""
			for name, _ := range v.specTypes[pipeline.ResultMetaName] {
				names += name + " "
			}
			return fmt.Errorf("结果表列数不符，请核对：%v", names)
		}
	}
	return nil
}

// 检查表的可达和冗余
func (v *validator) validateTopology(pipeline *pb.Pipeline) ([]*pb.Collection, error) {
	var unreachables []*pb.Collection
	for _, c := range pipeline.Collections {
		unreachables = append(unreachables, c)
	}
	used := make(map[string]bool)
	reachableCollections := make(map[string]bool)
	var sequence []*pb.Collection
	for len(unreachables) > 0 {
		var newUnreachables []*pb.Collection
		for _, c := range unreachables {
			reachable := false
			switch body := c.Body.(type) {
			case *pb.Collection_Input:
				reachable = true
			case *pb.Collection_Projection:
				if reachableCollections[body.Projection.Input] {
					used[body.Projection.Input] = true
					reachable = true
				}
			case *pb.Collection_Filter:
				if reachableCollections[body.Filter.Input] {
					used[body.Filter.Input] = true
					reachable = true
				}
			case *pb.Collection_Join:
				if reachableCollections[body.Join.LeftInput] && reachableCollections[body.Join.RightInput] {
					used[body.Join.LeftInput] = true
					used[body.Join.RightInput] = true
					reachable = true
				}
			case *pb.Collection_Group:
				if reachableCollections[body.Group.Input] {
					used[body.Group.Input] = true
					reachable = true
				}
			default:
				panic("表类型未知")
			}
			if reachable {
				reachableCollections[c.Name] = true
				sequence = append(sequence, c)
			} else {
				newUnreachables = append(newUnreachables, c)
			}
		}
		if len(unreachables) == len(newUnreachables) {
			return nil, fmt.Errorf("不可达的表 %v", unreachables[0].Name)
		}
		unreachables = newUnreachables
	}
	output := 0
	for _, c := range pipeline.Collections {
		if !used[c.Name] {
			output++
		}
	}
	if output != 1 {
		return nil, fmt.Errorf("必须有恰好一个表作为最终输出")
	}
	return sequence, nil
}

func (v *validator) validateType(c *pb.Collection) (map[string]pb.Type, error) {
	if cached, existed := v.fieldType[c.Name]; existed {
		return cached, nil
	}

	fieldType := make(map[string]pb.Type)
	switch body := c.Body.(type) {
	case *pb.Collection_Input:
		specType, ok := v.specTypes[body.Input.MetaName]
		if !ok {
			return nil, fmt.Errorf("输入表 %v 引用的表模板 %v 不存在", c.Name, body.Input.MetaName)
		}
		for k, v := range specType {
			fieldType[k] = v
		}
	case *pb.Collection_Filter:
		subType, err := v.validateType(v.collectionByName[body.Filter.Input])
		if err != nil {
			return nil, err
		}
		condType, err := v.validateExpression(body.Filter.Expression, subType, fmt.Sprintf("表 %v", c.Name))
		if err != nil {
			return nil, err
		}
		if condType != pb.Type_BOOLEAN {
			return nil, fmt.Errorf("表 %v 的过滤表达式类型需为真/假", c.Name)
		}
		for k, v := range subType {
			fieldType[k] = v
		}
	case *pb.Collection_Projection:
		subType, err := v.validateType(v.collectionByName[body.Projection.Input])
		if err != nil {
			return nil, err
		}
		for _, field := range body.Projection.Fields {
			if _, existed := fieldType[field.Name]; existed {
				return nil, fmt.Errorf("表 %v 有多个名为 %v 的列", c.Name, field.Name)
			}
			t, err := v.validateExpression(field.Expression, subType, fmt.Sprintf("表 %v 的列 %v", c.Name, field.Name))
			if err != nil {
				return nil, err
			}
			fieldType[field.Name] = t
		}
	case *pb.Collection_Join:
		leftType, err := v.validateType(v.collectionByName[body.Join.LeftInput])
		if err != nil {
			return nil, err
		}
		rightType, err := v.validateType(v.collectionByName[body.Join.RightInput])
		if err != nil {
			return nil, err
		}
		if len(body.Join.Conditions) == 0 {
			return nil, fmt.Errorf("表 %v 的连接条件为空", c.Name)
		}
		for index, cond := range body.Join.Conditions {
			lType, ok := leftType[cond.Left]
			if !ok {
				return nil, fmt.Errorf("表 %v 的第 %v 个过滤条件中引用了不存在的左列 %v", c.Name, index+1, cond.Left)
			}
			rType, ok := rightType[cond.Right]
			if !ok {
				return nil, fmt.Errorf("表 %v 的第 %v 个过滤条件中引用了不存在的右列 %v", c.Name, index+1, cond.Right)
			}
			if lType != rType {
				return nil, fmt.Errorf("表 %v 的第 %v 个过滤条件中左右两列类型不符", c.Name, index+1)
			}
		}
		for _, field := range body.Join.LeftFields {
			lType, ok := leftType[field.Field]
			if !ok {
				return nil, fmt.Errorf("表 %v 引用了不存在的左列 %v", c.Name, field.Field)
			}
			if _, existed := fieldType[field.Name]; existed {
				return nil, fmt.Errorf("表 %v 有多个名为 %v 的列", c.Name, field.Name)
			}
			fieldType[field.Name] = lType
		}
		for _, field := range body.Join.RightFields {
			rType, ok := rightType[field.Field]
			if !ok {
				return nil, fmt.Errorf("表 %v 引用了不存在的右列 %v", c.Name, field.Field)
			}
			if _, existed := fieldType[field.Name]; existed {
				return nil, fmt.Errorf("表 %v 有多个名为 %v 的列", c.Name, field.Name)
			}
			fieldType[field.Name] = rType
		}
	case *pb.Collection_Group:
		subType, err := v.validateType(v.collectionByName[body.Group.Input])
		if err != nil {
			return nil, err
		}
		hasKey := make(map[string]bool)
		for _, key := range body.Group.Keys {
			if _, ok := subType[key]; !ok {
				return nil, fmt.Errorf("表 %v 引用了不存在列 %v", c.Name, key)
			}
			if hasKey[key] {
				return nil, fmt.Errorf("表 %v 有多个名为 %v 的分组列", c.Name, key)
			}
			hasKey[key] = true
		}
		for _, field := range body.Group.Fields {
			if _, existed := fieldType[field.Name]; existed {
				return nil, fmt.Errorf("表 %v 有多个名为 %v 的列", c.Name, field.Name)
			}
			t, err := v.validateExpression(field.Expression, subType, fmt.Sprintf("表 %v 的列 %v", c.Name, field.Name))
			if err != nil {
				return nil, err
			}
			fieldType[field.Name] = t
			isAgg, _, err := v.expressionAggregation(field.Expression, hasKey, fmt.Sprintf("表 %v 的列 %v", c.Name, field.Name))
			if err != nil {
				return nil, err
			}
			if !isAgg {
				return nil, fmt.Errorf("表 %v 的列 %v 未被聚合", c.Name, field.Name)
			}
		}
	default:
		panic("表类型未知")
	}

	v.fieldType[c.Name] = fieldType
	return fieldType, nil
}

func (v *validator) validateExpression(expr *pb.Expression, subType map[string]pb.Type, loc string) (pb.Type, error) {
	switch body := expr.Body.(type) {
	case *pb.Expression_Field:
		if t, ok := subType[body.Field]; ok {
			return t, nil
		} else {
			return pb.Type_FLOAT, fmt.Errorf("%v 公式中引用了不存在的字段 %v", loc, body.Field)
		}
	case *pb.Expression_Literal:
		switch body.Literal.Body.(type) {
		case *pb.Literal_BooleanValue:
			return pb.Type_BOOLEAN, nil
		case *pb.Literal_FloatValue:
			return pb.Type_FLOAT, nil
		case *pb.Literal_StringValue:
			return pb.Type_STRING, nil
		default:
			panic("未知常量类型")
		}
	case *pb.Expression_Operation:
		return v.validateOperation(body.Operation, subType, loc)
	default:
		panic("未知表达式类型")
	}
}

func (v *validator) validateOperation(oper *pb.Operation, subType map[string]pb.Type, loc string) (pb.Type, error) {
	switch oper.Operator {
	case pb.Operator_ADD, pb.Operator_MULTIPLY:
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_FLOAT {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是实数", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_FLOAT, nil

	case pb.Operator_SUBTRACT, pb.Operator_DIVIDE, pb.Operator_MOD, pb.Operator_POW:
		if len(oper.Operands) != 2 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有两个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_FLOAT {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是实数", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_FLOAT, nil

	case pb.Operator_EXP, pb.Operator_SQRT, pb.Operator_LN, pb.Operator_LOG2, pb.Operator_LOG10, pb.Operator_ABS, pb.Operator_CEIL, pb.Operator_TRUNC, pb.Operator_FLOOR:
		if len(oper.Operands) != 1 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有一个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_FLOAT {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是实数", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_FLOAT, nil

	case pb.Operator_EQ, pb.Operator_NE:
		if len(oper.Operands) != 2 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有两个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		var randType []pb.Type
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			randType = append(randType, t)
		}
		if randType[0] != randType[1] {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算中两个操作数类型不符", loc, pb.Operator_name[int32(oper.Operator)])
		}
		return pb.Type_BOOLEAN, nil

	case pb.Operator_LT, pb.Operator_LTE, pb.Operator_GT, pb.Operator_GTE:
		if len(oper.Operands) != 2 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有两个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_FLOAT {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是实数", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_BOOLEAN, nil

	case pb.Operator_EXISTS:
		if len(oper.Operands) != 1 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有一个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		if field, ok := oper.Operands[0].Body.(*pb.Expression_Field); ok {
			if _, ok := subType[field.Field]; ok {
				return pb.Type_BOOLEAN, nil
			} else {
				return pb.Type_FLOAT, fmt.Errorf("%v 公式中引用了不存在的字段 %v", loc, field.Field)
			}
		} else {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中\"存在\"运算符的操作数必须是一个字段", loc)
		}

	case pb.Operator_AND, pb.Operator_OR:
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_BOOLEAN {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是真/假类型", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_BOOLEAN, nil

	case pb.Operator_NOT:
		if len(oper.Operands) != 1 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有一个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_BOOLEAN {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是真/假类型", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_BOOLEAN, nil

	case pb.Operator_COND:
		if len(oper.Operands) != 3 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 \"条件\" 运算的必须恰好有三个操作数", loc)
		}
		var randType []pb.Type
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			randType = append(randType, t)
		}
		if randType[0] != pb.Type_BOOLEAN {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 \"条件\" 运算的第一个操作数必须为真/假类型", loc)
		}
		if randType[1] != randType[2] {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 \"条件\" 运算的后两个操作类型不一致", loc)
		}
		return randType[1], nil

	case pb.Operator_REGEX:
		if len(oper.Operands) != 2 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有两个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		if lit, ok := oper.Operands[0].Body.(*pb.Expression_Literal); ok {
			if _, ok := lit.Literal.Body.(*pb.Literal_StringValue); ok {
				t, err := v.validateExpression(oper.Operands[1], subType, loc)
				if err != nil {
					return pb.Type_FLOAT, err
				}
				if t != pb.Type_STRING {
					return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 \"正则\" 运算的第二个操作数必须为字符串类型", loc)
				}
				return pb.Type_BOOLEAN, nil
			}
		}
		return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 \"正则\" 运算的第一个操作数必须为常量字符串", loc)
	case pb.Operator_SUM, pb.Operator_AVG, pb.Operator_MAX, pb.Operator_MIN:
		if len(oper.Operands) != 1 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有一个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
			if t != pb.Type_FLOAT {
				return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的操作数不是实数", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_FLOAT, nil
	case pb.Operator_STRING_AGG:
		if len(oper.Operands) != 1 {
			return pb.Type_STRING, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有一个操作数11", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			t, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_STRING, err
			}
			if t != pb.Type_STRING {
				return pb.Type_STRING, fmt.Errorf("%v 的公式中 %v 运算的操作数不是实数11", loc, pb.Operator_name[int32(oper.Operator)])
			}
		}
		return pb.Type_STRING, nil
	case pb.Operator_COUNT, pb.Operator_YEAR, pb.Operator_MONTH, pb.Operator_DAY:
		if len(oper.Operands) != 1 {
			return pb.Type_FLOAT, fmt.Errorf("%v 的公式中 %v 运算的必须恰好有一个操作数", loc, pb.Operator_name[int32(oper.Operator)])
		}
		for _, rand := range oper.Operands {
			_, err := v.validateExpression(rand, subType, loc)
			if err != nil {
				return pb.Type_FLOAT, err
			}
		}
		return pb.Type_FLOAT, nil
	case pb.Operator_CONCATE:
		if len(oper.Operands) <= 1 {
			return pb.Type_STRING, fmt.Errorf("至少有两个元素进行连接")
		}
		return pb.Type_STRING, nil
	}

	panic("未知的运算符")
}

func (v *validator) expressionAggregation(expression *pb.Expression, hasKey map[string]bool, loc string) (bool, bool, error) {
	switch body := expression.Body.(type) {
	case *pb.Expression_Field:
		return hasKey[body.Field], !hasKey[body.Field], nil
	case *pb.Expression_Literal:
		return true, true, nil
	case *pb.Expression_Operation:
		return v.operationAggregation(body.Operation, hasKey, loc)
	}
	panic("未知表达式类型")
}

func (v *validator) operationAggregation(operation *pb.Operation, hasKey map[string]bool, loc string) (bool, bool, error) {
	switch operation.Operator {
	case pb.Operator_SUM, pb.Operator_AVG, pb.Operator_MAX, pb.Operator_MIN, pb.Operator_COUNT, pb.Operator_STRING_AGG:
		_, isNonAgg, err := v.expressionAggregation(operation.Operands[0], hasKey, loc)
		if err != nil {
			return false, false, err
		}
		if !isNonAgg {
			return false, false, fmt.Errorf("%v 中 %v 运算符的操作数不可被聚集", loc, pb.Operator_name[int32(operation.Operator)])
		}
		return true, false, nil
	default:
		agg, nonAgg := true, true
		for _, rand := range operation.Operands {
			isAgg, isNonAgg, err := v.expressionAggregation(rand, hasKey, loc)
			if err != nil {
				return false, false, err
			}
			if !isAgg {
				agg = false
			}
			if !isNonAgg {
				nonAgg = false
			}
		}
		return agg, nonAgg, nil
	}
}
