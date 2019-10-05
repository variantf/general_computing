import React from "react";
import { hashHistory } from "react-router";
import {
  listDatabaseTable,
  upsertDatabaseTable,
  getTableMetadata
} from "./webapi";
import { TYPE_TEXT } from "./types";
import styles from "./data-import.css";

class DatabaseTableList extends React.Component {
  constructor(props) {
    super(props);
    const { name, path } = this.props.location.query;
    this.state = {
      tableMetadata: { path: path, name: name, hint: "", fields: [] }, // 原始TableMetadata
      simpleMetaFields: {}, // 二维列简化后的TableMetadata
      databaseTables: [], // 原始DatabaseTable列表
      editingTable: { metaPath: path, metaName: name, dbName: "", fields: [] } // 正在编辑的原始DatabaseTable
    };
  }
  async componentDidMount() {
    try {
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  async refresh() {
    const { name, path } = this.props.location.query;
    let tableMetadata = await getTableMetadata(path, name);
    let simpleMetaFields = this.calcSimpleMetaFieldMap(tableMetadata);
    let editingTable = {
      metaPath: path,
      metaName: name,
      dbName: "",
      fields: []
    };
    for (let name of Object.keys(simpleMetaFields)) {
      editingTable.fields.push({
        stdName: name,
        alias: [name, simpleMetaFields[name].hint]
      });
    }
    this.setState({
      tableMetadata: tableMetadata,
      simpleMetaFields: simpleMetaFields,
      databaseTables: await listDatabaseTable(path, name),
      editingTable: editingTable
    });
  }
  async editDatabaseTable(databaseTable) {
    this.setState({
      // deep copy strings
      editingTable: {
        metaPath: databaseTable.metaPath,
        metaName: databaseTable.metaName,
        dbName: databaseTable.dbName,
        fields: databaseTable.fields
      }
    });
  }
  handleDbNameChange(e) {
    this.state.editingTable.dbName = e.target.value;
    this.forceUpdate();
  }
  handleFieldAliasChange(stdName, idx, e) {
    const { editingTable } = this.state;
    this.setState({
      editingTable: {
        ...editingTable,
        fields: editingTable.fields.map(f => ({
          ...f,
          alias: f.alias.map((a, i) => {
            if (f.stdName == stdName && i == idx) {
              return e.target.value;
            }
            return a;
          })
        }))
      }
    });
  }
  calcSimpleMetaFieldMap(tableMetadata) {
    let res = {};
    for (let field of tableMetadata.fields) {
      let name = field.name;
      let pos = name.lastIndexOf("[");
      let ewb = 0;
      if (name[name.length - 1] == "]" && pos > 0) {
        // 二维列
        let num = parseInt(name.substring(pos + 1, field.name.length - 1));
        ewb = Math.max(ewb, num);
        name = name.substring(0, pos);
      }
      res[name] = {
        type: field.type,
        hint: field.hint,
        ewb: ewb
      };
    }
    console.log("simple meta: ", res);
    return res;
  }
  checkUpsert(dbTable, tableMeta) {
    if (
      dbTable.metaPath != tableMeta.path ||
      dbTable.metaName != tableMeta.name ||
      dbTable.dbName == ""
    ) {
      return "实际数据表名不能为空";
    }
    let simpleMetaFields = this.calcSimpleMetaFieldMap(tableMeta);
    if (Object.keys(simpleMetaFields).length != dbTable.fields.length)
      return "列映射数量不完整";
    for (let field of dbTable.fields) {
      let arr = field.alias;
      for (let each of arr) {
        if (each.length <= 0) return "列别名不能为空";
      }
      for (let other of dbTable.fields) {
        if (field.stdName != other.stdName) {
          let arr2 = other.alias;
          for (let s1 of arr)
            for (let s2 of arr2)
              if (s1 == s2) return "不同的标准列存在相同的别名: " + s1;
        }
      }
    }
    return "ok";
  }
  async handleUpsert() {
    const { editingTable, tableMetadata } = this.state;
    try {
      let checkMsg = this.checkUpsert(editingTable, tableMetadata);
      if (checkMsg != "ok") {
        alert("格式错误：" + checkMsg);
        return;
      }
      console.log("upsert:", editingTable);
      await upsertDatabaseTable(editingTable);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  handleAliasDeletion = (stdName, idx) => {
    const { editingTable } = this.state;
    this.setState({
      editingTable: {
        ...editingTable,
        fields: editingTable.fields.map(f => ({
          ...f,
          alias: f.alias.filter((a, i) => f.stdName != stdName || i != idx)
        }))
      }
    });
  };

  handleAliasAddition = stdName => {
    const { editingTable } = this.state;
    this.setState({
      editingTable: {
        ...editingTable,
        fields: editingTable.fields.map(f => {
          let newField = { ...f };
          if (f.stdName == stdName) {
            newField.alias.push("");
          }
          return newField;
        })
      }
    });
  };
  render() {
    const { editingTable, tableMetadata, simpleMetaFields } = this.state;
    let editingFieldsMap = Object.assign(
      ...editingTable.fields.map(field => ({
        [field.stdName]: { stdName: field.stdName, alias: field.alias }
      })),
      {}
    );
    return (
      <div>
        <h4>已有“实际数据表”：</h4>
        <p>
          这些实际数据表(DatabaseTable)绑定的标准表信息(TableMetadata)：<br />
          <b>{tableMetadata.path}: {tableMetadata.name}</b>
          &nbsp;
          {tableMetadata.hint}
        </p>
        <p>
          点击“上传”选择要插入数据的目标，并跳转到上传Excel/csv页面。<br />
          点击“编辑”查看或更新与标准表信息之间的列映射。<br />
          按下Ctrl+F查找。
        </p>
        <div style={{ maxHeight: "20em", overflowY: "scroll" }}>
          <ul>
            {this.state.databaseTables.map((dbTable, index) => (
              <li key={index}>
                {dbTable.dbName}
                &nbsp;
                <a
                  href="javascript:"
                  onClick={() =>
                    hashHistory.push({
                      pathname: "/upload",
                      query: { dbName: dbTable.dbName }
                    })}
                >
                  上传
                </a>
                &nbsp;
                <a
                  href="javascript:"
                  onClick={this.editDatabaseTable.bind(this, dbTable)}
                >
                  编辑
                </a>
              </li>
            ))}
          </ul>
        </div>
        <hr />
        <h4 id="edit">编辑“实际表数据”：</h4>
        <p>
          <input
            type="text"
            value={editingTable.dbName}
            onChange={this.handleDbNameChange.bind(this)}
            placeholder="实际数据表名"
            className="large"
          />&nbsp;
        </p>
        <table>
          <thead>
            <tr>
              <th>标准列名称</th>
              <th>标准列类型</th>
              <th>标准列备注</th>
              <th>二维长度(0为一维)</th>
              <th>实际列别名(可有多个候选,逗号分隔)</th>
            </tr>
          </thead>
          <tbody>
            {Object.keys(simpleMetaFields).map(name => {
              return (
                <tr key={name}>
                  <td>{name}</td>
                  <td>{TYPE_TEXT[simpleMetaFields[name].type]}</td>
                  <td>{simpleMetaFields[name].hint}</td>
                  <td>{simpleMetaFields[name].ewb}</td>
                  <td>
                    {editingFieldsMap[name].alias.map((a, idx) => (
                      <span key={"alias" + idx}>
                        <input
                          type="text"
                          value={a}
                          onChange={this.handleFieldAliasChange.bind(
                            this,
                            name,
                            idx
                          )}
                          className="large"
                        />
                        <button
                          onClick={this.handleAliasDeletion.bind(
                            this,
                            name,
                            idx
                          )}
                        >
                          删
                        </button>
                      </span>
                    ))}
                    <button onClick={this.handleAliasAddition.bind(this, name)}>
                      加
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <button type="button" onClick={this.handleUpsert.bind(this)}>提交</button>
      </div>
    );
  }
}

export default DatabaseTableList;
