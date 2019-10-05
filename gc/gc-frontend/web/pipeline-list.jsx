import React from "react";
import { hashHistory } from "react-router";
import {
  getPipelines,
  addPipeline,
  deletePipeline,
  getTasks,
  addTask,
  deleteTask,
  duplicatePipeline,
  recalculateEnabledTasks,
  fetchDBConnections,
  removeDBConnection,
  addDBConnection
} from "./webapi";

class PipelineList extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      pipelines: [],
      tasks: []
    };
    this.handleAdd = this.handleAdd.bind(this);
    this.handleAddTask = this.handleAddTask.bind(this);
    this.handleRecalculateEnabled = this.handleRecalculateEnabled.bind(this);
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
    this.setState({ pipelines: await getPipelines(), tasks: await getTasks(), connections: await fetchDBConnections() });
  }

  async duplicatePipeline(pipeline) {
    let path = prompt("请输入路径名称", "");
    let name = prompt("请输入公式名称", "");
    if (name == null || name == "" || path == null || path == "") {
      return;
    }
    try {
      await duplicatePipeline(pipeline.path, pipeline.name, path, name);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败: ", msg.data);
    }
  }
  async removePipeline(pipeline) {
    try {
      await deletePipeline(pipeline.path, pipeline.name);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  async removeTask(task) {
    try {
      await deleteTask(task.path, task.name);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  async handleAdd() {
    try {
      await addPipeline(this.refs.name.value, this.refs.path.value);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  async handleAddTask() {
    try {
      await addTask(this.refs.taskName.value, this.refs.taskPath.value);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败：" + msg.data);
    }
  }
  async handleRecalculateEnabled(){
    try {
      await recalculateEnabledTasks(this.refs.alltaskPath.value);
      await this.refresh();
    } catch (msg) {
      console.log(msg);
      alert('失败：' + msg.data);
    }
  }
  addDBConnection = async () => {
    const {db_conn_name, username, password, ipadd, port, instance} = this.refs;

    try {
      await addDBConnection(db_conn_name.value, username.value, 
        password.value, ipadd.value, port.value, instance.value) 
      this.refresh();
    } catch (msg) {
      console.log(msg);
      alert("失败" + msg.data);
    }
  }
  render() {
    return (
      <div>
        <h4>导入数据</h4>
        数据导入流程：<br />
        <ol>
          <li>选择 “标准表模板” (TableMetadata)。</li>
          <li>建立 “实际数据表” (DatabaseTable)，并确定列名映射关系。</li>
          <li>上传 Excel 或 csv 表格，由服务器进行解析。解析成功--&gt;插入数据，解析失败--&gt;修正重来。</li>
        </ol>
        <a href="javascript:" onClick={()=> hashHistory.push({pathname: '/tablemetadata'})}>开始导入</a>
        <hr/>
        <h4>添加数据库信息</h4>
        名称：<input type="text" ref="db_conn_name" placeholder="任意名称" />
        用户名：<input type="text" ref="username" placeholder="数据库用户名" />
        密码：<input type="password" ref="password" placeholder="数据库密码" />
        地址：<input type="text" ref="ipadd" placeholder="数据库IP地址" />
        端口：<input type="text" ref="port" placeholder="数据库端口" />
        实例名：<input type="text" ref="instance" placeholder="实例名" />
        <br />
        <button onClick={this.addDBConnection}>添加</button>
        {this.state.connections? this.state.connections.map((conn, idx) => {
          return (<div key={"conn"+idx}>{conn.name}&nbsp;&nbsp;<a href="javascript:" onClick={async ()=>{
            await removeDBConnection(conn.name);
            this.refresh()
          }}>删</a></div>)
        }): null}

        <hr/>
        <table><tr><td style={{verticalAlign: "top"}}>
        <h4 id="exist-pipeline">已有公式（按下Ctrl+F查找）：</h4>
        <ul>
          {this.state.pipelines.map((pipeline, index) =>
            <li key={index}>
              <a href="javascript:" onClick={()=> hashHistory.push({pathname: '/pipeline', query:{name:pipeline.name, path: pipeline.path}})}>
                {pipeline.path}: {pipeline.name}
              </a>
              &nbsp;
              <a href="javascript:" onClick={this.duplicatePipeline.bind(this, pipeline)}>复制</a>
              &nbsp;
              <a href="javascript:" onClick={this.removePipeline.bind(this, pipeline)}>删</a>
            </li>
          )}
        </ul>
        <h4>添加公式</h4>
        <input type="text" ref="name" placeholder="公式名称" />
        <input type="text" ref="path" placeholder="公式路径" />
        <button type="button" onClick={this.handleAdd}>添加</button>
        </td><td style={{verticalAlign: "top"}}>
        <h4 id="exist-task">已有任务（按下Ctrl+F查找）：</h4>
        <input type="text" ref="alltaskPath" placeholder="任务路径" />
        <button type="button" onClick={this.handleRecalculateEnabled}>重新计算</button>
        
        <ul>
          {this.state.tasks.map((task, index) =>
            <li key={index}>
              <a href="javascript:" onClick={()=>hashHistory.push({pathname: '/task', query:{name:task.name, path:task.path}})}>
                {task.path}: {task.name}
              </a>
              &nbsp;
              <a href="javascript:" onClick={this.removeTask.bind(this, task)}>删</a>
            </li>
          )}
        </ul>
        <h4>添加任务</h4>
        <input type="text" ref="taskName" placeholder="任务名称" />
        <input type="text" ref="taskPath" placeholder="任务路径" />
        <button type="button" onClick={this.handleAddTask}>添加</button>
        </td></tr></table>
      </div>
    );
  }
}

export default PipelineList;