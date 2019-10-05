import store from "./store";

export const FORMULA_TYPE_TEXT = {
  SAVE_ONLY: "测试版",
  ACTIVE: "正式版"
};

export const TYPE_TEXT = {
  FLOAT: "实数",
  STRING: "字串",
  BOOLEAN: "真/假",
  DATETIME: "日期时间"
};

export const COLLECTION_TYPE_TEXT = {
  INPUT: "输入",
  SAMPLE: "取样",
  PROJECTION: "计算",
  FILTER: "过滤",
  JOIN: "连接",
  GROUP: "分组"
};

export const OPERATOR_TEXT = {
  ADD: "+",
  SUBTRACT: "-",
  MULTIPLY: "×",
  DIVIDE: "÷",
  MOD: "求余数",
  POW: "乘方",
  EXP: "e的次方",
  SQRT: "根号",
  LN: "ln",
  LOG2: "log2",
  LOG10: "log10",
  ABS: "绝对值",
  CEIL: "上取整",
  TRUNC: "向零取整",
  FLOOR: "下取整",
  EQ: "=",
  NE: "≠",
  LT: "<",
  LTE: "≤",
  GT: ">",
  GTE: "≥",
  EXISTS: "存在",
  AND: "与",
  OR: "或",
  NOT: "非",
  COND: "条件",
  REGEX: "正则",
  SUM: "求和(聚合)",
  AVG: "平均值(聚合)",
  MAX: "最大值(聚合)",
  MIN: "最小值(聚合)",
  COUNT: "数量(聚合)",
  CONCATE: "字符串连接",
  YEAR: "年份",
  MONTH: "月份",
  DAY: "日期",
  STRING_AGG:"字符(聚合)"
};

export function collectionType(collection) {
  if ("input" in collection) return "INPUT";
  if ("sample" in collection) return "SAMPLE";
  if ("projection" in collection) return "PROJECTION";
  if ("filter" in collection) return "FILTER";
  if ("join" in collection) return "JOIN";
  if ("group" in collection) return "GROUP";
  console.log("非法集合", collection);
  throw "非法集合";
}

export function collectionFields(collection, path) {
  let visited = new Set();
  for (;;) {
    if (collection == null) return [];
    if (visited.has(collection.name)) return null;
    visited.add(collection.name);

    const type = collectionType(collection);
    if (type == "INPUT") {
      let metas = store.getState().tableMetadata[path];
      if (!metas) return [];
      let meta = metas[collection.input.metaName];
      if (!meta) return [];
      return Array.from(meta.fields);
    } else if (type == "SAMPLE") collection = collection.sample.input;
    else if (type == "PROJECTION")
      return Array.from(collection.projection.fields);
    else if (type == "FILTER") collection = collection.filter.input;
    else if (type == "JOIN")
      return Array.from(collection.join.leftFields).concat(
        collection.join.rightFields
      );
    else if (type == "GROUP") return Array.from(collection.group.fields);
    else throw "非法集合";
  }
}

export function expressionType(expression) {
  if ("field" in expression) return "FIELD";
  if ("literal" in expression) return "LITERAL";
  if ("operation" in expression) return "OPERATION";
  throw "非法表达式";
}

export function literalType(literal) {
  if ("floatValue" in literal) return "FLOAT";
  if ("stringValue" in literal) return "STRING";
  if ("booleanValue" in literal) return "BOOLEAN";
  throw "非法常数值";
}

export function objToPipeline(pipeline) {
  pipeline.collections = pipeline.collections || [];
  let clMap = {};
  for (let c of pipeline.collections) {
    clMap[c.name] = {};
  }
  return {
    collections: pipeline.collections.map(c =>
      objFillCollection(c, clMap[c.name], clMap)
    ),
    path: pipeline.path,
    name: pipeline.name,
    resultMetaName: pipeline.resultMetaName,
    taxType: pipeline.taxType,
    industryCode: pipeline.industryCode,
    industryName: pipeline.industryName,
    tags: pipeline.tags
  };
}

export function pipelineToObj(pipeline) {
  return {
    collections: pipeline.collections.map(collectionToObj),
    path: pipeline.path,
    name: pipeline.name,
    resultMetaName: pipeline.resultMetaName,
    tax_type: pipeline.tax_type,
    industry_code: pipeline.industry_code,
    industry_name: pipeline.industry_name,
    tags: pipeline.tags
  };
}

export function objFillCollection(collection, result, clMap) {
  result.name = collection.name;
  result.output = collection.output;
  if ("input" in collection) {
    result.input = collection.input;
  } else if ("sample" in collection) {
    result.sample = objToSample(collection.sample, clMap);
  } else if ("projection" in collection) {
    result.projection = objToProjection(collection.projection, clMap);
  } else if ("filter" in collection) {
    result.filter = objToFilter(collection.filter, clMap);
  } else if ("join" in collection) {
    result.join = objToJoin(collection.join, clMap);
  } else if ("group" in collection) {
    result.group = objToGroup(collection.group, clMap);
  }
  return result;
}

function collectionToObj(collection) {
  let result = { name: collection.name, output: collection.output };
  if ("input" in collection) {
    result.input = collection.input;
  } else if ("sample" in collection) {
    result.sample = sampleToObj(collection.sample, collection.name);
  } else if ("projection" in collection) {
    result.projection = projectionToObj(collection.projection, collection.name);
  } else if ("filter" in collection) {
    result.filter = filterToObj(collection.filter, collection.name);
  } else if ("join" in collection) {
    result.join = joinToObj(collection.join, collection.name);
  } else if ("group" in collection) {
    result.group = groupToObj(collection.group, collection.name);
  }
  return result;
}

function sampleToObj(sample, loc) {
  let result = { rate: sample.rate };
  if (sample.input == null) throw `${loc}的被取样表为空`;
  result.input = sample.input.name;
  return result;
}

function objToSample(sample, clMap) {
  return {
    input: clMap[sample.input] || null,
    rate: sample.rate
  };
}

function filterToObj(filter, loc) {
  let result = {};
  if (filter.input == null) throw `${loc}的被过滤表为空`;
  result.input = filter.input.name;
  if (filter.expression == null) throw `${loc}的过滤表达式为空`;
  result.expression = expressionToObj(filter.expression);
  return result;
}

function objToFilter(filter, clMap) {
  return {
    input: clMap[filter.input] || null,
    expression: objToExpression(filter.expression)
  };
}

function projectionToObj(projection, loc) {
  let result = {};
  if (projection.input == null) throw `${loc}的源表为空`;
  result.input = projection.input.name;
  result.fields = projection.fields.map(field => {
    let location = `${loc}的${field.name}列`;
    let result = { name: field.name };
    if (field.expression == null) throw `${location}的表达式为空`;
    result.expression = expressionToObj(field.expression);
    return result;
  });
  return result;
}

function objToProjection(projection, clMap) {
  return {
    input: clMap[projection.input] || null,
    fields: (projection.fields || []).map(field => ({
      name: field.name,
      expression: objToExpression(field.expression)
    }))
  };
}

function expressionToObj(expression) {
  if ("literal" in expression)
    return { literal: literalToObj(expression.literal) };
  else if ("field" in expression) {
    return { field: expression.field.name };
  } else if ("operation" in expression) {
    return { operation: operationToObj(expression.operation) };
  }
}

function objToExpression(expression) {
  if ("literal" in expression)
    return { literal: objToLiteral(expression.literal) };
  else if ("field" in expression) {
    return { field: { name: expression.field } };
  } else if ("operation" in expression) {
    return { operation: objToOperation(expression.operation) };
  }
}

function literalToObj(literal) {
  return literal;
}

function objToLiteral(literal) {
  return literal;
}

function operationToObj(operation) {
  return {
    operator: operation.operator,
    operands: operation.operands.map(expressionToObj)
  };
}

function objToOperation(operation) {
  return {
    operator: operation.operator,
    operands: operation.operands.map(objToExpression)
  };
}

function joinToObj(join, loc) {
  let result = {
    method: join.method,
    conditions: join.conditions.map((condition, index) => {
      let result = {};
      if (condition.left == null) throw `${loc}的第${index + 1}个条件的左表列为空`;
      result.left = condition.left.name;
      if (condition.right == null) throw `${loc}的第${index + 1}个条件的右表列为空`;
      result.right = condition.right.name;
      return result;
    }),
    leftFields: join.leftFields.map((field, index) => {
      let result = {
        name: field.name
      };
      if (field.field == null) throw `${loc}的左表输出的第${index + 1}个列来源列为空`;
      result.field = field.field.name;
      return result;
    }),
    rightFields: join.rightFields.map((field, index) => {
      let result = {
        name: field.name
      };
      if (field.field == null) throw `${loc}的右表输出的第${index + 1}个列来源列为空`;
      result.field = field.field.name;
      return result;
    })
  };
  if (join.leftInput == null) throw `${loc}的左表为空`;
  result.leftInput = join.leftInput.name;
  if (join.rightInput == null) throw `${loc}的右表为空`;
  result.rightInput = join.rightInput.name;
  return result;
}

function objToJoin(join, clMap) {
  return {
    leftInput: clMap[join.leftInput],
    rightInput: clMap[join.rightInput],
    method: join.method,
    conditions: (join.conditions || []).map(condition => ({
      left: { name: condition.left },
      right: { name: condition.right }
    })),
    leftFields: (join.leftFields || []).map(field => ({
      name: field.name,
      field: { name: field.field }
    })),
    rightFields: (join.rightFields || []).map(field => ({
      name: field.name,
      field: { name: field.field }
    }))
  };
}

function objToGroup(group, clMap) {
  return {
    input: clMap[group.input],
    keys: (group.keys || []).map(key => ({ name: key })),
    fields: (group.fields || []).map(field => ({
      name: field.name,
      expression: objToExpression(field.expression)
    }))
  };
}

function groupToObj(group, loc) {
  let result = {
    keys: group.keys.map(f => f.name),
    fields: group.fields.map((field, index) => {
      let result = { name: field.name };
      if (field.expression == null) throw `${loc}的第${index + 1}个输出列的表达式为空`;
      result.expression = expressionToObj(field.expression);
      return result;
    })
  };
  if (group.input == null) throw `${loc}的源表为空`;
  result.input = group.input.name;
  return result;
}
