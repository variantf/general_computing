import React from "react";
import $ from "jquery";
import { hashHistory } from "react-router";
import {
  getDatabaseTable,
  fetchDataFileList,
  analyzeDataFile,
  fetchDBTables,
  fetchDBGroups,
  fetchDBConnections,
  LoadDataFromDB
} from "./webapi";
import { TYPE_TEXT } from "./types";

class UploadPage extends React.Component {
  constructor(props) {
    super(props);
    const { dbName } = this.props.location.query;
    this.state = {
      dbTable: { metaPath: "", metaName: "", dbName: dbName, fields: [] },
      files: [],
      result: ""
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
    const { dbName } = this.props.location.query;
    this.setState({
      dbTable: await getDatabaseTable(dbName),
      files: await fetchDataFileList(dbName),
      conns: await fetchDBConnections()
    });
    console.log(this.state.dbTable);
    console.log(this.state.files);
  }
  handleFileUpload(e) {
    e.preventDefault();
    if (this.refs.file.files.length < 1) {
      alert("请选择文件！");
      return;
    }
    let fd = new FormData();
    for (var i = 0; i < this.refs.file.files.length; i++) {
      fd.append("file", this.refs.file.files[i]);
    }
    fd.append("dbName", this.state.dbTable.dbName);
    this.setState({ result: "正在上传" });
    $.ajax({
      url: "http://192.168.143.238:12101/upload-file",
      context: this,
      data: fd,
      processData: false,
      contentType: false,
      type: "POST",
      success: function(data) {
        this.setState({ result: "上传成功：" + data });
        this.refresh();
      },
      error: function(xhr, err) {
        console.log(xhr, err);
        this.setState({
          result: "上传失败：" +
            xhr.status +
            " " +
            xhr.responseText +
            " " +
            err +
            " 参见浏览器Console"
        });
      }
    });
  }
  async handleAnalyze(fileName) {
    try {
      this.setState({ result: "正在分析：" + fileName });
      let result = await analyzeDataFile(this.state.dbTable.dbName, fileName);
      this.setState({ result: result });
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }

  handleDBLoad = async () => {
    const {
      currentConn,
      currGroup,
      currTable,
      dbTable: { dbName }
    } = this.state;
    try {
      this.setState({ result: "正在导入数据：" + dbName });
      const result = await LoadDataFromDB(
        currentConn,
        currGroup,
        currTable,
        dbName
      );
      this.setState({ result: result });
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  };
  dbConnChange = async event => {
    let conn = event.target.value;
    this.setState({ currentConn: conn, groups: await fetchDBGroups(conn) });
  };

  dbTableChange = async event => {
    let table = event.target.value;
    this.setState({ currTable: table });
  };

  dbGroupChange = async event => {
    let newGroup = event.target.value;
    this.setState({
      tables: await fetchDBTables(this.state.currentConn, newGroup),
      currGroup: newGroup
    });
  };
  render() {
    const { dbTable, files, result } = this.state;
    console.log(this.state.conns, this.state.groups, this.state.tables);
    return (
      <div>
        <h4>上传数据</h4>
        <div>
          <p><b>实际数据表名：</b>{dbTable.dbName} </p>
          <p><b>标准表名：</b>{dbTable.metaName} </p>
          <p><b>路径：</b>{dbTable.metaPath} </p>
          <table>
            <thead>
              <tr>
                <th>标准列名</th>
                <th>实际列别名</th>
              </tr>
            </thead>
            <tbody>
              {dbTable.fields.map(field => (
                <tr key={field.stdName}>
                  <td>{field.stdName}</td>
                  <td>{field.alias}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <hr />
        <form encType="multipart/form-data">
          <p>
            <input
              ref="file"
              type="file"
              name="file"
              accept="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, text/csv"
              multiple
            />
          </p>
          <p>支持xlsx/csv，csv文件请使用UTF-8编码，暂不支持xls文件，请手动用Excel转换为xlsx。</p>
          <input
            type="button"
            value="上传并保存表格"
            onClick={this.handleFileUpload.bind(this)}
          />
        </form>
        <h4>已上传表格</h4>
        <ul>
          {files.map((f, index) => (
            <li key={index}>
              {f} &nbsp;
              <a href="javascript:" onClick={this.handleAnalyze.bind(this, f)}>
                分析并插入数据
              </a>{" "}
              &nbsp;
            </li>
          ))}
        </ul>
        <p>{result}</p>
        <h4>从数据库导入</h4>
        数据库连接<select onChange={this.dbConnChange}>
          <option>请选择</option>
          {this.state.conns
            ? this.state.conns.map(conn => (
                <option key={conn.name}>{conn.name}</option>
              ))
            : null}
        </select>
        <br />
        数据表租<select onChange={this.dbGroupChange}>
          <option>请选择</option>
          {this.state.groups
            ? this.state.groups.map(group => (
                <option key={group}>{group}</option>
              ))
            : null}
        </select>
        <br />
        数据表<select onChange={this.dbTableChange}>
          <option>请选择</option>
          {this.state.tables
            ? this.state.tables
                .filter(
                  table =>
                    table.indexOf(this.refs.table_filter.value || "") != -1
                )
                .map(table => <option key={table}>{table}</option>)
            : null}
        </select>
        <input
          ref="table_filter"
          placeholder="筛选表名"
          onChange={event => {
            this.setState({ table_filter: event.target.value });
          }}
        />
        <br />
        <button onClick={this.handleDBLoad}>导入</button>
      </div>
    );
  }
}

export default UploadPage;
