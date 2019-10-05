import React from 'react';
import {connect} from 'react-redux';
import {loadDatabases} from './actions';
import {getPipelines, getTask, getPipeline, updateTask, debugTask,fetchTestResult,fetchRunTime} from './webapi';

class TaskEdit extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      pipelines: [],
      pipeline: null,
      task: null,
      testresult:"",
      runtime:"",
      filter: "",
      append_: false
    };
    this.handleChangePipeline = this.handleChangePipeline.bind(this);
    this.handleChangeEnabled = this.handleChangeEnabled.bind(this);
    this.handleChangeAppend = this.handleChangeAppend.bind(this);
    this.handleSave = this.handleSave.bind(this);
    this.handleDebug = this.handleDebug.bind(this);
  }
  async componentDidMount() {
    const {path, name} = this.props.location.query;
    let task = await getTask(path, name);
    task.inputMapping = task.inputMapping || [];
    this.setState({
      pipelines: (await getPipelines()).filter(p=>p.path == path),
      task: task
    });
    try {
      await this.refreshPipeline(task.pipelineName);
    }catch(msg){
    }
    this.handleFetchTestResult();
    this.handleFetchRunTime();
  }
  async handleChangePipeline(e) {
    this.state.task.pipelineName = e.target.value;
    this.forceUpdate();
    await this.refreshPipeline(e.target.value);
  }
  async refreshPipeline(name) {
    let pipeline = await getPipeline(this.state.task.path, name);
    let index = 0;
    for (let c of pipeline.collections) {
      if (c.input) {
        this.state.task.inputMapping[index] = this.state.task.inputMapping[index] || {};
        this.state.task.inputMapping[index].collectionName = c.name;
        this.props.loadDatabases(pipeline.path, c.input.metaName);
        index++;
      }
    }
    this.props.loadDatabases(pipeline.path, pipeline.resultMetaName);
    this.state.task.inputMapping.length = index;
    this.setState({
      pipeline
    });
  }
  handleChangeEnabled(e) {
    this.state.task.enabled = e.target.checked;
    this.forceUpdate();
  }
  handleChangeDatabaseName(index, e) {
    this.state.task.inputMapping[index].databaseName = e.target.value;
    this.forceUpdate();
  }
  handleChangeOutputDatabaseName(e) {
    this.state.task.outputMapping = e.target.value;
    this.forceUpdate();
  }
  handleChangeAppend(e) {
    this.state.append_ = e.target.checked;
    this.forceUpdate();
  }
  handleFilterChange = e => {
    this.setState({ filter: e.target.value });
  };
  async handleSave() {
    try {
      let inputs = this.state.pipeline.collections.filter(c => c.input);
      for (let i in inputs) {
        if (!this.state.task.inputMapping[i].databaseName) {
          let c = inputs[i];
          let metas = this.props.databases[c.input.metaName];
          if (metas && metas.length >= 1) {
            this.state.task.inputMapping[i].databaseName = metas[0].dbName;
          }
        }
      }
      await updateTask(this.state.task, this.state.append_);
      alert('保存成功');
    } catch (msg) {
      console.log(msg);
      alert('错误：' + msg.data);
    }
  }
  async handleFetchTestResult(){
    const {path, name} = this.props.location.query;
    try {
      let json = await fetchTestResult(path, name);       
      if (json.testresult == "") {
        this.setState({testresult:"还没测试运算过"})        
      }else{
        this.setState({testresult:json.testresult})
      }    
    } catch (msg) {
      console.log(msg);
      alert('失败：' + msg.data);
    }
  }
  async handleFetchRunTime(){
    const {path, name} = this.props.location.query;
    try {
      let json = await fetchRunTime(path, name);  
      if (json.runtime == "") {
        this.setState({runtime:"未计算"})        
      }else{
        this.setState({runtime:json.runtime}) 
      }

    } catch (msg) {
      console.log(msg);
      alert('失败：' + msg.data);
    }
  }
  async refresh() {
    this.setState({ pipelines: await getPipelines()});
  }
  async handleDebug() {
    const {path, name} = this.props.location.query;
    try {
      let json = await debugTask(path, name);
      // let newWin = window.open();
      // if (!newWin) {
      //   alert('请允许弹出窗口！');
      //   return;
      // }
      // let document = newWin.document;
      document.title = '运算结果';
      let pre = document.createElement('pre');
      pre.innerHTML = json;
      document.body.appendChild(pre);
      // await debugTask(path, name);
      await this.refresh();
      this.handleFetchTestResult();
    } catch (msg) {
      console.log(msg);
      alert('出错了:' + msg.data);
    }
  }
  render() {
    
    const {path, name} = this.props.location.query;
    const {task, pipelines, pipeline,testresult,runtime} = this.state;
    if (task == null) return <h1>正在加载，请稍候</h1>;
    let inputs = [];
    let output = "";
    if (pipeline) {
      inputs = pipeline.collections.filter(c => c.input);
      output = pipeline.resultMetaName;
    }
    console.log(output);
    let filter = this.state.filter;
    let options = [<option selected />];
    for (let i in pipelines) {
      if (pipelines[i].name.indexOf(filter) != -1) {
        options.push(<option key={i} value={pipelines[i].name}>{pipelines[i].name}</option>);
      }
    }
    
    return (
      <div>
        <p><a href="javascript" href="/">返回首页</a></p>
        <h4>编辑任务 {path}: {name}</h4>
        <p>
          公式：
          <input
            type="text"
            size="100"
            onChange={this.handleFilterChange}
            value={filter}
            placeholder="过滤，留空则显示全部"
          />
          <select value={task.pipelineName} onChange={this.handleChangePipeline}>
            {options}
          </select>
        </p>
        <p>
          是否投入正式计算：
          <input type="checkbox" checked={task.enabled} onChange={this.handleChangeEnabled} />
        </p>
        <p>
          <table>
            <thead>
              <tr>
                <th>输入表</th>
                <th>数据库表</th>
              </tr>
            </thead>
            <tbody>
              {inputs.map((input, index) =>
                <tr key={index}>
                  <td>{input.name}</td>
                  <td>
                    {task.inputMapping[index] &&
                    <select value={task.inputMapping[index].databaseName} onChange={this.handleChangeDatabaseName.bind(this, index)}>
                      <option selected disabled>请选择</option>
                      {(this.props.databases[input.input.metaName] || []).map(db =>
                        <option key={db.dbName} value={db.dbName}>{db.dbName}</option>
                      )}
                    </select>}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </p>
        <p>
          <table>
            <thead>
              <tr>
                <th>输出表</th>
                <th>数据库表</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>{output}</td>
                <td>
                  <select value={task.outputMapping} onChange={this.handleChangeOutputDatabaseName.bind(this)}>
                    <option key="" value="" selected disabled>请选择</option>
                    {(this.props.databases[output] || []).map(db =>
                      <option key={db.dbName} value={db.dbNamename}>{db.dbName}</option>
                    )}
                  </select>
                </td>
              </tr>
            </tbody>
          </table>
        </p>
        {/*
        <p>
          是否保留旧结果（不删除原有结果，直接追加）：
          <input type="checkbox" checked={this.state.append_} onChange={this.handleChangeAppend} />
        </p>
        */}
        <p>
          <button type="button" onClick={this.handleSave}>保存</button>
          <button type="button" onClick={this.handleDebug}>测试运行</button>
        </p>
       <p>上次运行时间：{runtime}</p>
        <div dangerouslySetInnerHTML={{__html: testresult}}>
          
        </div>
      </div>
    );
  }
}

export default connect((state, ownProps) => {
  return {databases: state.databases[ownProps.location.query.path] || {}};
}, {loadDatabases})(TaskEdit);