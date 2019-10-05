import React from "react";
import { hashHistory } from "react-router";
import {
  listTableMetadata,
  upsertTableMetadata,
  getTableMetadata,
  deleteTableMetadata
} from "./webapi";
import { TYPE_TEXT } from "./types";
import styles from "./data-import.css";

class TableMetadataList extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      metadatas: [],
      editingMetadata: { path: "", name: "", hint: "", fields: [] }
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
    this.setState({
      metadatas: await listTableMetadata("", true),
      editingMetadata: { path: "", name: "", hint: "", fields: [] }
    });
  }
  async editMetadata(metadata) {
    this.setState({
      editingMetadata: await getTableMetadata(metadata.path, metadata.name)
    });
  }
  handlePathChange(e) {
    this.state.editingMetadata.path = e.target.value;
    this.forceUpdate();
  }
  handleNameChange(e) {
    this.state.editingMetadata.name = e.target.value;
    this.forceUpdate();
  }
  handleHintChange(e) {
    this.state.editingMetadata.hint = e.target.value;
    this.forceUpdate();
  }
  handleAddField() {
    this.state.editingMetadata.fields.push({
      name: "XL",
      type: "FLOAT",
      hint: "新列"
    });
    this.forceUpdate();
  }
  handleClearField() {
    this.state.editingMetadata.fields = [];
    this.forceUpdate();
  }
  handleFieldNameChange(index, e) {
    const names = e.target.value.split("\n");
    const { fields } = this.state.editingMetadata;
    for (let i = 0; i < names.length; i++) {
      const idx = index + i;
      if (fields.length > idx) {
        fields[idx].name = names[i];
      } else {
        fields[idx] = {
          name: names[i],
          type: "FLOAT",
          hint: "新列"
        };
      }
    }
    this.forceUpdate();
  }
  handleFieldTypeChange(index, e) {
    this.state.editingMetadata.fields[index].type = e.target.value;
    this.forceUpdate();
  }
  handleFieldHintChange(index, e) {
    const hints = e.target.value.split("\n");
    const { fields } = this.state.editingMetadata;
    console.log("fields.length", fields.length);
    for (let i = 0; i < hints.length; i++) {
      const idx = index + i;
      if (fields.length > idx) {
        fields[idx].hint = hints[i];
      } else {
        fields[idx] = {
          name: "XL",
          type: "FLOAT",
          hint: hints[i]
        };
      }
    }
    this.forceUpdate();
  }
  removeField(index) {
    this.state.editingMetadata.fields.splice(index, 1);
    this.forceUpdate();
  }
  async removeMetadata(metadata) {
    try {
      await deleteTableMetadata(metadata.path, metadata.name);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  async handleUpsert() {
    const { editingMetadata } = this.state;
    try {
      if (editingMetadata.path == "" || editingMetadata.name == "") {
        return;
      }
      console.log("upsert:", editingMetadata);
      await upsertTableMetadata(editingMetadata);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  render() {
    const { editingMetadata } = this.state;
    return (
      <div>
        <h4>已有“标准表信息”：</h4>
        <p>
          点击“表名”进行选择，并跳转到实际数据表映射页面。<br />
          点击“编辑”查看或更新标准表信息。<br />
          按下Ctrl+F查找。
        </p>
        <div style={{ maxHeight: "30em", overflowY: "scroll" }}>
          <ul>
            {this.state.metadatas.map((metadata, index) => (
              <li key={index}>
                {metadata.path}: &nbsp;
                <a
                  href="javascript:"
                  onClick={() =>
                    hashHistory.push({
                      pathname: "/databasetable",
                      query: { name: metadata.name, path: metadata.path }
                    })}
                >
                  {metadata.name}
                </a>
                &nbsp;
                {metadata.hint}
                &nbsp;
                <a
                  href="javascript:"
                  onClick={this.editMetadata.bind(this, metadata)}
                >
                  编辑
                </a>
                &nbsp;
                <a
                  href="javascript:"
                  onClick={this.removeMetadata.bind(this, metadata)}
                >
                  删除
                </a>
              </li>
            ))}
          </ul>
        </div>
        <hr />
        <h4 id="edit">编辑“标准表信息”：</h4>
        <div>
          <input
            type="text"
            value={editingMetadata.path}
            onChange={this.handlePathChange.bind(this)}
            placeholder="路径"
          />&nbsp;
          <input
            type="text"
            value={editingMetadata.name}
            onChange={this.handleNameChange.bind(this)}
            placeholder="名称（可填写拼音表名）"
            className="large"
          />&nbsp;
          <input
            type="text"
            value={editingMetadata.hint}
            onChange={this.handleHintChange.bind(this)}
            placeholder="备注（可填写中文表名）"
            className="large"
          />&nbsp;
        </div>
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>类型</th>
              <th>备注</th>
              <th>
                <a href="javascript:" onClick={this.handleAddField.bind(this)}>
                  加
                </a>&nbsp;
                <a
                  href="javascript:"
                  onClick={this.handleClearField.bind(this)}
                >
                  清空
                </a>
              </th>
            </tr>
          </thead>
          <tbody>
            {editingMetadata.fields &&
              editingMetadata.fields.map((field, index) => (
                <tr key={index}>
                  <td>
                    <textarea
                      type="text"
                      value={field.name}
                      onChange={this.handleFieldNameChange.bind(this, index)}
                    />
                  </td>
                  <td>
                    <select
                      value={field.type}
                      onChange={this.handleFieldTypeChange.bind(this, index)}
                    >
                      {Object.keys(TYPE_TEXT).map(t => (
                        <option key={t} value={t}>
                          {TYPE_TEXT[t]}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td>
                    <textarea
                      type="text"
                      value={field.hint}
                      onChange={this.handleFieldHintChange.bind(this, index)}
                      className="large"
                    />
                  </td>
                  <td>
                    <a
                      href="javascript:"
                      onClick={this.removeField.bind(this, index)}
                    >
                      删
                    </a>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
        <button type="button" onClick={this.handleUpsert.bind(this)}>
          提交
        </button>
      </div>
    );
  }
}

export default TableMetadataList;
