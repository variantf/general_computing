import React from "react";
import { connect } from "react-redux";
import { loadTableMetadata } from "./actions";
import Collection from "./collection";
import CollectionHolder from "./collection-holder";
import Field from "./field";
import FieldHolder from "./field-holder";
import ExpressionHolder from "./expression-holder";
import {
  TYPE_TEXT,
  COLLECTION_TYPE_TEXT,
  collectionType,
  collectionFields
} from "./types";
import styles from "./collection-editor.css";

class CollectionEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleNameChange = this.handleNameChange.bind(this);
    this.handleOutputChange = this.handleOutputChange.bind(this);
  }
  handleNameChange(e) {
    this.props.collection.name = e.target.value;
    this.forceUpdate();
  }
  handleOutputChange(output) {
    this.props.handleOutputChange(output);
  }
  render() {
    const { collection, path, output } = this.props;
    if (collection == null)
      return (
        <div>
          <h2>设定输出模板</h2>
          <OutputEditor
            path={path}
            output={output}
            handleOutputChange={this.handleOutputChange}
          />
        </div>
      );
    const type = collectionType(collection);
    let editor;
    if (type == "INPUT")
      editor = <InputEditor path={path} input={collection.input} />;
    else if (type == "SAMPLE")
      editor = <SampleEditor sample={collection.sample} />;
    else if (type == "PROJECTION")
      editor = (
        <ProjectionEditor path={path} projection={collection.projection} />
      );
    else if (type == "FILTER")
      editor = <FilterEditor filter={collection.filter} />;
    else if (type == "JOIN")
      editor = <JoinEditor path={path} join={collection.join} />;
    else if (type == "GROUP")
      editor = <GroupEditor path={path} group={collection.group} />;
    return (
      <div>
        <h2>
          {collection.name}
          <span className={styles.type}>{COLLECTION_TYPE_TEXT[type]}</span>
        </h2>
        <div>
          名称：
          <input
            type="text"
            value={collection.name}
            onChange={this.handleNameChange}
          />
        </div>
        {editor}
      </div>
    );
  }
}
export default CollectionEditor;

class InputEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleNameChange = this.handleNameChange.bind(this);
    this.state = {
      filter: ""
    };
  }
  componentDidMount() {
    this.props.loadTableMetadata(this.props.path);
  }
  handleNameChange(e) {
    this.props.input.metaName = e.target.value;
    this.forceUpdate();
  }
  handleFilterChange = e => {
    this.setState({ filter: e.target.value });
  };
  render() {
    const { tableMetadata } = this.props;
    let input_collection = this.props.input;
    let filter = this.state.filter;
    console.log(tableMetadata, input_collection);
    if (!tableMetadata) return null;
    let options = [<option selected />];
    for (let name in tableMetadata) {
      if (name.indexOf(filter) != -1) {
        options.push(<option key={name} value={name}>{name}</option>);
      }
    }
    let metadata = tableMetadata[input_collection.metaName] || { fields: [] };

    return (
      <div>
        <div>
          表模板：
          <input
            type="text"
            size="100"
            onChange={this.handleFilterChange}
            value={filter}
            placeholder="过滤，留空则显示全部"
          />
          <select
            value={input_collection.metaName}
            onChange={this.handleNameChange}
          >
            {options}
          </select>
        </div>
        <div>列：</div>
        <ul>
          {metadata.fields.map(field => (
            <li key={field.name}>{field.name}: {TYPE_TEXT[field.type]}</li>
          ))}
        </ul>
      </div>
    );
  }
}

InputEditor = connect(
  (state, ownProps) => {
    return { tableMetadata: state.tableMetadata[ownProps.path] };
  },
  { loadTableMetadata }
)(InputEditor);

class SampleEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleRateChange = this.handleRateChange.bind(this);
    this.handleInputChange = this.handleInputChange.bind(this);
  }
  handleRateChange(e) {
    this.props.sample.rate = Number(e.target.value);
    this.forceUpdate();
  }
  handleInputChange(collection) {
    this.props.sample.input = collection;
    this.forceUpdate();
  }
  render() {
    const { sample } = this.props;
    return (
      <div>
        <div>
          被取样表：
          <CollectionHolder
            collection={sample.input}
            onChange={this.handleInputChange}
          />
        </div>
        <div>
          取样率：
          <input
            type="number"
            style={{ width: 100 }}
            size="10"
            min="0"
            max="1"
            step="0.0001"
            value={sample.rate}
            onChange={this.handleRateChange}
          />
          0到1之间。0表示不取样，1表示全部取样。
        </div>
      </div>
    );
  }
}

class FilterEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleExpressionChange = this.handleExpressionChange.bind(this);
  }
  handleInputChange(collection) {
    this.props.filter.input = collection;
    this.forceUpdate();
  }
  handleExpressionChange(expression) {
    this.props.filter.expression = expression;
    this.forceUpdate();
  }
  render() {
    const { filter } = this.props;
    return (
      <div>
        <div>
          被过滤表：
          <CollectionHolder
            collection={filter.input}
            onChange={this.handleInputChange}
          />
        </div>
        <div>
          过滤表达式：
          <ExpressionHolder
            needConfirm
            expression={filter.expression}
            onChange={this.handleExpressionChange}
          />
          仅保留此表达式为“真”的数据项
        </div>
      </div>
    );
  }
}

class ProjectionEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleAddOneField = this.handleAddOneField.bind(this);
    this.handleAddFields = this.handleAddFields.bind(this);
    this.handleClearField = this.handleClearField.bind(this);
  }
  handleInputChange(collection) {
    this.props.projection.input = collection;
    this.forceUpdate();
  }
  handleClearField() {
    this.props.projection.fields = [];
    this.forceUpdate();
  }
  handleAddOneField() {
    this.props.projection.fields.push({ name: "新列", expression: null });
    this.forceUpdate();
  }
  handleAddFields() {
    const fields = collectionFields(
      this.props.projection.input,
      this.props.path
    );
    if (fields == null) {
      alert("当前输入表存在循环引用，请消除循环引用后重试。");
      return;
    }
    for (let field of fields) {
      this.props.projection.fields.push({
        name: field.name,
        expression: {
          field: field
        }
      });
    }
    this.forceUpdate();
  }
  handleFieldNameChange(index, e) {
    this.props.projection.fields[index].name = e.target.value;
    this.forceUpdate();
  }
  handleFieldExpressionChange(index, expression) {
    this.props.projection.fields[index].expression = expression;
    this.forceUpdate();
  }
  removeField(index) {
    this.props.projection.fields.splice(index, 1);
    this.forceUpdate();
  }
  render() {
    const { projection } = this.props;
    return (
      <div>
        <div>
          源表：
          <CollectionHolder
            collection={projection.input}
            onChange={this.handleInputChange}
          />
          {projection.input &&
            <a
              href="javascript:"
              className={styles.op}
              onClick={this.handleAddFields}
            >
              导入所有列
            </a>}
        </div>
        <table>
          <caption>输出列</caption>
          <thead>
            <tr>
              <th>列名</th>
              <th>公式</th>
              <th>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleAddOneField}
                >
                  加
                </a>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleClearField}
                >
                  清空
                </a>
              </th>
            </tr>
          </thead>
          <tbody>
            {projection.fields.map((field, index) => (
              <tr key={index}>
                <td>
                  <input
                    type="text"
                    value={field.name}
                    onChange={this.handleFieldNameChange.bind(this, index)}
                  />
                </td>
                <td>
                  <ExpressionHolder
                    needConfirm
                    expression={field.expression}
                    onChange={this.handleFieldExpressionChange.bind(
                      this,
                      index
                    )}
                  />
                </td>
                <td>
                  <a
                    href="javascript:"
                    className={styles.op}
                    onClick={() => this.removeField(index)}
                  >
                    删
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }
}

class JoinEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleChangeLeftInput = this.handleChangeLeftInput.bind(this);
    this.handleChangeRightInput = this.handleChangeRightInput.bind(this);
    this.handleChangeMethod = this.handleChangeMethod.bind(this);
    this.handleAddCond = this.handleAddCond.bind(this);
    this.handleAddLeftField = this.handleAddLeftField.bind(this);
    this.handleClearLeftField = this.handleClearLeftField.bind(this);
    this.handleAddRightField = this.handleAddRightField.bind(this);
    this.handleClearRightField = this.handleClearRightField.bind(this);
    this.handleImportLeftFields = this.handleImportLeftFields.bind(this);
    this.handleImportRightFields = this.handleImportRightFields.bind(this);
  }
  handleChangeLeftInput(collection) {
    this.props.join.leftInput = collection;
    this.forceUpdate();
  }
  handleChangeRightInput(collection) {
    this.props.join.rightInput = collection;
    this.forceUpdate();
  }
  handleChangeMethod(e) {
    this.props.join.method = e.target.value;
    this.forceUpdate();
  }
  handleAddCond() {
    this.props.join.conditions.push({
      left: null,
      right: null
    });
    this.forceUpdate();
  }
  handleAddLeftField() {
    this.props.join.leftFields.push({
      name: "新列",
      field: null
    });
    this.forceUpdate();
  }
  handleAddRightField() {
    this.props.join.rightFields.push({
      name: "新列",
      field: null
    });
    this.forceUpdate();
  }
  handleClearLeftField() {
    this.props.join.leftFields = [];
    this.forceUpdate();
  }
  handleClearRightField() {
    this.props.join.rightFields = [];
    this.forceUpdate();
  }
  handleImportLeftFields(index) {
    const fields = collectionFields(this.props.join.leftInput, this.props.path);
    if (fields == null) {
      alert("该表目前存在循环引用，请先消除循环引用。");
      return;
    }
    for (let field of fields) {
      this.props.join.leftFields.push({
        name: field.name,
        field
      });
    }
    this.forceUpdate();
  }
  handleImportRightFields(index) {
    const fields = collectionFields(
      this.props.join.rightInput,
      this.props.path
    );
    if (fields == null) {
      alert("该表目前存在循环引用，请先消除循环引用。");
      return;
    }
    for (let field of fields) {
      this.props.join.rightFields.push({
        name: field.name,
        field
      });
    }
    this.forceUpdate();
  }
  handleCondLeftFieldChange(cond, field) {
    cond.left = field;
    this.forceUpdate();
  }
  handleCondRightFieldChange(cond, field) {
    cond.right = field;
    this.forceUpdate();
  }
  handleFieldNameChange(field, e) {
    field.name = e.target.value;
    this.forceUpdate();
  }
  handleFieldChange(field, f) {
    field.field = f;
    this.forceUpdate();
  }
  removeCond(index) {
    this.props.join.conditions.splice(index, 1);
    this.forceUpdate();
  }
  removeLeftField(index) {
    this.props.join.leftFields.splice(index, 1);
    this.forceUpdate();
  }
  removeRightField(index) {
    this.props.join.rightFields.splice(index, 1);
    this.forceUpdate();
  }
  render() {
    const { join } = this.props;
    return (
      <div>
        <div>
          左源表：
          <CollectionHolder
            collection={join.leftInput}
            onChange={this.handleChangeLeftInput}
          />
          {!join.leftInput ||
            <a
              href="javascript:"
              className={styles.op}
              onClick={this.handleImportLeftFields}
            >
              导入所有列
            </a>}
        </div>
        <div>
          右源表：
          <CollectionHolder
            collection={join.rightInput}
            onChange={this.handleChangeRightInput}
          />
          {!join.rightInput ||
            <a
              href="javascript:"
              className={styles.op}
              onClick={this.handleImportRightFields}
            >
              导入所有列
            </a>}
        </div>
        <div>
          连接方式：
          <select value={join.method} onChange={this.handleChangeMethod}>
            <option value="INNER">不能缺数据</option>
            <option value="LEFT">左表不可缺数据</option>
            <option value="RIGHT">右表不可缺数据</option>
            <option value="FULL">两表都可缺数据</option>
          </select>
        </div>
        <table className={styles.table}>
          <caption>连接条件</caption>
          <thead>
            <tr>
              <th>左表列</th>
              <th>条件</th>
              <th>右表列</th>
              <th>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleAddCond}
                >
                  加
                </a>
              </th>
            </tr>
          </thead>
          <tbody>
            {join.conditions.map((cond, index) => (
              <tr key={index}>
                <td>
                  <FieldHolder
                    field={cond.left}
                    onChange={this.handleCondLeftFieldChange.bind(this, cond)}
                  />
                </td>
                <td>=</td>
                <td>
                  <FieldHolder
                    field={cond.right}
                    onChange={this.handleCondRightFieldChange.bind(this, cond)}
                  />
                </td>
                <td>
                  <a
                    href="javascript:"
                    className={styles.op}
                    onClick={this.removeCond.bind(this, index)}
                  >
                    删
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <table className={styles.table}>
          <caption>左表输出列</caption>
          <thead>
            <tr>
              <th>名称</th>
              <th>来源列</th>
              <th>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleAddLeftField}
                >
                  加
                </a>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleClearLeftField}
                >
                  清空
                </a>
              </th>
            </tr>
          </thead>
          <tbody>
            {join.leftFields.map((field, index) => (
              <tr key={index}>
                <td>
                  <input
                    type="text"
                    value={field.name}
                    onChange={this.handleFieldNameChange.bind(this, field)}
                  />
                </td>
                <td>
                  <FieldHolder
                    field={field.field}
                    onChange={this.handleFieldChange.bind(this, field)}
                  />
                </td>
                <td>
                  <a
                    href="javascript:"
                    className={styles.op}
                    onClick={this.removeLeftField.bind(this, index)}
                  >
                    删
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <table className={styles.table}>
          <caption>右表输出列</caption>
          <thead>
            <tr>
              <th>名称</th>
              <th>来源列</th>
              <th>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleAddRightField}
                >
                  加
                </a>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleClearRightField}
                >
                  清空
                </a>
              </th>
            </tr>
          </thead>
          <tbody>
            {join.rightFields.map((field, index) => (
              <tr key={index}>
                <td>
                  <input
                    type="text"
                    value={field.name}
                    onChange={this.handleFieldNameChange.bind(this, field)}
                  />
                </td>
                <td>
                  <FieldHolder
                    field={field.field}
                    onChange={this.handleFieldChange.bind(this, field)}
                  />
                </td>
                <td>
                  <a
                    href="javascript:"
                    className={styles.op}
                    onClick={this.removeRightField.bind(this, index)}
                  >
                    删
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }
}

class GroupEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleChangeInput = this.handleChangeInput.bind(this);
    this.handleAddKey = this.handleAddKey.bind(this);
    this.handleAddField = this.handleAddField.bind(this);
  }
  handleChangeInput(collection) {
    this.props.group.input = collection;
    this.forceUpdate();
  }
  handleAddKey(field) {
    this.props.group.keys.push(field);
    this.forceUpdate();
  }
  handleRemoveKey(index) {
    this.props.group.keys.splice(index, 1);
    this.forceUpdate();
  }
  handleAddField() {
    this.props.group.fields.push({
      name: "新列",
      expression: null
    });
    this.forceUpdate();
  }
  handleChangeFieldName(index, e) {
    this.props.group.fields[index].name = e.target.value;
    this.forceUpdate();
  }
  handleChangeFieldExpression(index, expression) {
    this.props.group.fields[index].expression = expression;
    this.forceUpdate();
  }
  handleRemoveField(index) {
    this.props.group.fields.splice(index, 1);
    this.forceUpdate();
  }
  render() {
    const { group } = this.props;
    return (
      <div>
        <div>
          源表：
          <CollectionHolder
            collection={group.input}
            onChange={this.handleChangeInput}
          />
        </div>
        <div>分组列：</div>
        <ul>
          {group.keys.map((key, index) => (
            <li key={index}>
              <Field field={key} />
              <a
                href="javascript:"
                onClick={this.handleRemoveKey.bind(this, index)}
                className={styles.op}
              >
                删
              </a>
            </li>
          ))}
        </ul>
        <div><FieldHolder onChange={this.handleAddKey} /></div>
        <table className={styles.table}>
          <caption>输出列</caption>
          <thead>
            <tr>
              <th>名称</th>
              <th>公式</th>
              <th>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.handleAddField}
                >
                  加
                </a>
              </th>
            </tr>
          </thead>
          <tbody>
            {group.fields.map((field, index) => (
              <tr key={index}>
                <td>
                  <input
                    type="text"
                    value={field.name}
                    onChange={this.handleChangeFieldName.bind(this, index)}
                  />
                </td>
                <td>
                  <ExpressionHolder
                    needConfirm
                    expression={field.expression}
                    onChange={this.handleChangeFieldExpression.bind(
                      this,
                      index
                    )}
                  />
                </td>
                <td>
                  <a
                    href="javascript:"
                    className={styles.op}
                    onClick={this.handleRemoveField.bind(this, index)}
                  >
                    删
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }
}

class OutputEditor extends React.Component {
  constructor(props) {
    super(props);
    this.handleNameChange = this.handleNameChange.bind(this);
  }
  componentDidMount() {
    this.props.loadTableMetadata(this.props.path);
  }
  handleNameChange(e) {
    this.props.handleOutputChange(e.target.value);
  }
  render() {
    const { tableMetadata, output } = this.props;
    if (!tableMetadata) return null;
    let options = [];
    options.push(<option key="请选择" value="">请选择</option>);
    for (let name in tableMetadata) {
      options.push(<option key={name} value={name}>{name}</option>);
    }
    let metadata = tableMetadata[output] || { fields: [] };
    return (
      <div>
        <div>
          表模板：
          <select value={output} onChange={this.handleNameChange}>
            {options}
          </select>
        </div>
        <div>列：</div>
        <ul>
          {metadata.fields.map(field => (
            <li key={field.name}>{field.name}: {TYPE_TEXT[field.type]}</li>
          ))}
        </ul>
      </div>
    );
  }
}

OutputEditor = connect(
  (state, ownProps) => {
    return { tableMetadata: state.tableMetadata[ownProps.path] };
  },
  { loadTableMetadata }
)(OutputEditor);
