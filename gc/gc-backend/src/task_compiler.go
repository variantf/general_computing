package server

import (
	"fmt"
	pb "git.corp.angel-salon.com/gc/proto"
	_ "github.com/lib/pq"
	"strings"
)

func quoteIdentifier(identifier string) string {
	return "`" + identifier + "`"
}

func ParseExpression(exp *pb.Expression) string {
	if literal := exp.GetLiteral(); literal != nil {
		if str, ok := literal.Body.(*pb.Literal_StringValue); ok {
			return "'" + strings.Replace(str.StringValue, "'", "''", -1) + "'"
		} else if num, ok := literal.Body.(*pb.Literal_FloatValue); ok {
			return fmt.Sprint(num.FloatValue)
		} else if boolean, ok := literal.Body.(*pb.Literal_BooleanValue); ok {
			if boolean.BooleanValue {
				return "true"
			} else {
				return "false"
			}
		}
		panic("Unknow Literal")
	} else if field, ok := exp.Body.(*pb.Expression_Field); ok {
		return quoteIdentifier(field.Field)
	} else if oper := exp.GetOperation(); oper != nil {
		exp_sql := ""
		switch oper.Operator {
		case pb.Operator_ABS:
			exp_sql = "ABS(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_ADD:
			exps := []string{}
			for _, exp := range oper.Operands {
				exps = append(exps, ParseExpression(exp))
			}
			exp_sql = "(" + strings.Join(exps, " + ") + ")"
		case pb.Operator_AND:
			exps := []string{}
			for _, exp := range oper.Operands {
				exps = append(exps, ParseExpression(exp))
			}
			exp_sql = "(" + strings.Join(exps, " and ") + ")"
		case pb.Operator_AVG:
			exp_sql = "AVG(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_CEIL:
			exp_sql = "CEIL(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_COND:
			exp_sql = "CASE WHEN " + ParseExpression(oper.GetOperands()[0]) + " THEN " +
				ParseExpression(oper.GetOperands()[1]) + " ELSE " + ParseExpression(oper.GetOperands()[2]) + " END"
		case pb.Operator_CONCATE:
			exps := []string{}
			for _, exp := range oper.Operands {
				exps = append(exps, ParseExpression(exp))
			}
			exp_sql = "(" + strings.Join(exps, " || ") + ")"
		case pb.Operator_COUNT:
			exp_sql = "COUNT(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_DAY:
			exp_sql = "EXTRACT(day from (timestamp '1970-01-01 00:00:00' + (" + ParseExpression(oper.GetOperands()[0]) + ") * interval '1 second'))"
		case pb.Operator_DIVIDE:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " / " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_EQ:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + /*"is not distinct from "*/ " = " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_EXISTS:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " IS NOT NULL)"
		case pb.Operator_EXP:
			exp_sql = "EXP(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_FLOOR:
			exp_sql = "FLOOR(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_GT:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " > " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_GTE:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " >= " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_LN:
			exp_sql = "FLOOR(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_LOG10:
			exp_sql = "LOG(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_LOG2:
			exp_sql = "LOG(" + ParseExpression(oper.GetOperands()[0]) + ", 2)"
		case pb.Operator_LT:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " < " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_LTE:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " <= " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_MAX:
			exp_sql = "MAX(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_MIN:
			exp_sql = "MIN(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_MOD:
			exp_sql = "MOD(" + ParseExpression(oper.GetOperands()[0]) + " , " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_MONTH:
			exp_sql = "EXTRACT(month from (timestamp '1970-01-01 00:00:00' + (" + ParseExpression(oper.GetOperands()[0]) + ") * interval '1 second'))"
		case pb.Operator_MULTIPLY:
			exps := []string{}
			for _, exp := range oper.Operands {
				exps = append(exps, ParseExpression(exp))
			}
			exp_sql = "(" + strings.Join(exps, " * ") + ")"
		case pb.Operator_NE:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + /*"is distinct from "*/ " <> " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_NOT:
			exp_sql = "NOT(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_OR:
			exps := []string{}
			for _, exp := range oper.Operands {
				exps = append(exps, ParseExpression(exp))
			}
			exp_sql = "(" + strings.Join(exps, " or ") + ")"
		case pb.Operator_POW:
			exp_sql = "POWER(" + ParseExpression(oper.GetOperands()[0]) + " , " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_REGEX:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[1]) + " SIMILAR TO " + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_SQRT:
			exp_sql = "SQRT(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_SUBTRACT:
			exp_sql = "(" + ParseExpression(oper.GetOperands()[0]) + " - " + ParseExpression(oper.GetOperands()[1]) + ")"
		case pb.Operator_SUM:
			exp_sql = "SUM(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_TRUNC:
			exp_sql = "TRUNC(" + ParseExpression(oper.GetOperands()[0]) + ")"
		case pb.Operator_YEAR:
			exp_sql = "EXTRACT(year from (timestamp '1970-01-01 00:00:00' + (" + ParseExpression(oper.GetOperands()[0]) + ") * interval '1 second'))"
		case pb.Operator_STRING_AGG:
			exp_sql = "string_agg( " + ParseExpression(oper.GetOperands()[0]) + ",',')"
		default:
			panic("Unknow operator")
		}
		return exp_sql
	} else {
		panic("Unknow expression")
	}
}

func ParseProjection(projection *pb.Projection) string {
	field_exps := []string{}
	for _, field := range projection.Fields {
		exp_sql := ParseExpression(field.Expression)
		field_exps = append(field_exps, exp_sql+" as "+quoteIdentifier(field.Name))
	}
	return "SELECT " + strings.Join(field_exps, ",") + " FROM " + quoteIdentifier(projection.Input)
}

func ParseFilter(filter *pb.Filter) string {
	return "SELECT * FROM " + quoteIdentifier(filter.Input) + "\nWHERE " +
		ParseExpression(filter.Expression)
}

func ParseGroup(group *pb.Group) string {
	field_exps := []string{}
	for _, field := range group.Fields {
		exp_sql := ParseExpression(field.Expression)
		field_exps = append(field_exps, exp_sql+" as "+quoteIdentifier(field.Name))
	}
	for idx, _ := range group.Keys {
		group.Keys[idx] = quoteIdentifier(group.Keys[idx])
	}
	return "SELECT " + strings.Join(field_exps, ",") + " FROM " +
		quoteIdentifier(group.Input) + " GROUP BY " +
		strings.Join(group.Keys, ",")
}

func ParseJoin(join *pb.Join) string {
	field_exps := []string{}
	for _, field := range join.LeftFields {
		exp_sql := "LeftTable." + quoteIdentifier(field.Field)
		field_exps = append(field_exps, exp_sql+" as "+quoteIdentifier(field.Name))
	}
	for _, field := range join.RightFields {
		exp_sql := "RightTable." + quoteIdentifier(field.Field)
		field_exps = append(field_exps, exp_sql+" as "+quoteIdentifier(field.Name))
	}
	method_str := ""
	switch join.Method {
	case pb.Join_FULL:
		method_str = "FULL OUTER"
	case pb.Join_INNER:
		method_str = "INNER"
	case pb.Join_LEFT:
		method_str = "LEFT"
	case pb.Join_RIGHT:
		method_str = "RIGHT"
	default:
		panic("Unsupported join method")
	}
	conditions := []string{}
	for _, cond := range join.Conditions {
		conditions = append(conditions,
			"LeftTable."+quoteIdentifier(cond.Left)+
				" = "+
				"RightTable."+quoteIdentifier(cond.Right))
	}
	return "SELECT " + strings.Join(field_exps, ",") + " FROM " +
		quoteIdentifier(join.LeftInput) + " as LeftTable " + method_str + " JOIN " +
		quoteIdentifier(join.RightInput) + " as RightTable ON " + strings.Join(conditions, " and ")
}

func (s *Server) CompileTask(task *pb.Task, pipeline *pb.Pipeline) (sql string, err error) {
	inputMapping := map[string]string{}
	for _, mapping := range task.InputMapping {
		inputMapping[mapping.CollectionName] = mapping.DatabaseName
	}
	sqlSequence := []string{}
	for _, collection := range pipeline.Collections {
		if input := collection.GetInput(); input != nil {

		} else if projection := collection.GetProjection(); projection != nil {
			if val, ok := inputMapping[projection.Input]; ok {
				projection.Input = val
			}
			select_sql := ParseProjection(projection)
			sqlSequence = append(sqlSequence, quoteIdentifier(collection.Name)+" as ("+select_sql+")")
		} else if filter := collection.GetFilter(); filter != nil {
			if val, ok := inputMapping[filter.Input]; ok {
				filter.Input = val
			}
			select_sql := ParseFilter(filter)
			sqlSequence = append(sqlSequence, quoteIdentifier(collection.Name)+" as ("+select_sql+")")
		} else if group := collection.GetGroup(); group != nil {
			if val, ok := inputMapping[group.Input]; ok {
				group.Input = val
			}
			select_sql := ParseGroup(group)
			sqlSequence = append(sqlSequence, quoteIdentifier(collection.Name)+" as ("+select_sql+")")
		} else if join := collection.GetJoin(); join != nil {
			if val, ok := inputMapping[join.LeftInput]; ok {
				join.LeftInput = val
			}
			if val, ok := inputMapping[join.RightInput]; ok {
				join.RightInput = val
			}
			select_sql := ParseJoin(join)
			sqlSequence = append(sqlSequence, quoteIdentifier(collection.Name)+" as ("+select_sql+")")
		} else {
			panic("Unknown Body Type")
		}
	}
	//lastName := pipeline.Collections[len(pipeline.Collections)-1].Name
	sql = "WITH " + strings.Join(sqlSequence, ",")
	return
}
