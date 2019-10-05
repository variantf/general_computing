
const fieldA = {name: '字段1', type: 'FLOAT'};
const fieldB = {name: '字段2', type: 'STRING'};
const fieldC = {name: '字段3', type: 'BOOLEAN'};
const fieldD = {name: '字段4', type: 'FLOAT'};

const inputA = {
  name: '输入表1',
  input: {
    metaName: '测试表1'
  }
};

const inputB = {
  name: '输入表2',
  input: {
    metaName: '测试表2'
  }
};

const sampleA = {
  name: '取样表1',
  sample: {
    input: inputA,
    rate: 0.8
  }
};

const projectionA = {
  name: '计算表1',
  projection: {
    input: null,
    fields: [
      {
        name: '计算列1',
        expression: {
          operation: {
            operator: 'ADD',
            operands: [
              {field: fieldA},
              {literal: {floatValue: 1}},
              {
                operation: {
                  operator: 'NOT',
                  operands: [{field: fieldC}]
                }
              }
            ]
          }
        }
      },
      {
        name: '字段3取反',
        expression: {
          operation: {
            operator: 'NOT',
            operands: [
              {field: fieldC}
            ]
          }
        }
      }
    ]
  }
}

const filterA = {
  name: '过滤表1',
  filter: {
    input: sampleA,
    expression: {literal: {floatValue: 1.23}}
  }
}

const joinA = {
  name: '连接表1',
  join: {
    leftInput: inputA,
    rightInput: inputB,
    conditions: [
      {left: fieldA, right: fieldC}
    ],
    leftFields: [
      {name: 'A的字段2', field: fieldB}
    ],
    rightFields: [
      {name: 'B的字段2', field: fieldD}
    ],
    method: 'LEFT'
  }
}

const groupA = {
  name: '分组表1',
  group: {
    input: inputA,
    keys: [fieldA],
    fields: [
      {name: '输出列A', expression: {field: fieldA}},
      {name: '输出列B', expression: {operation: {operator: 'SUM', operands:[{field: fieldB}]}}}
    ]
  }
}

export const examplePipeline = {
  collections: [inputA, inputB, sampleA, filterA, projectionA, joinA, groupA],
  path: '/风控',
  name: '测试公式'
};

export const metas = {
  '/风控': [
    {
      path: '/风控',
      name: '测试表1',
      fields: [fieldA, fieldB]
    },
    {
      path: '/风控',
      name: '测试表2',
      fields: [fieldC, fieldD]
    }
  ]
}