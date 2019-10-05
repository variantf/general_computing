import React from "react";
import { connect } from "react-redux";
import { hashHistory } from "react-router";
import { DragDropContext } from "react-dnd";
import HTML5Backend from "react-dnd-html5-backend";
import { loadTableMetadata } from "./actions";
import Collection from "./collection";
import CollectionEditor from "./collection-editor";
import FieldList from "./field-list";
import Expression from "./expression";
import styles from "./pipeline-edit.css";
import { pipelineToObj } from "./types";
import { getPipeline, updatePipeline, checkPipeline } from "./webapi";

class PipelineEdit extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      pipeline: null,
      editingCollection: null,
      showingFieldOf: null
    };
    this.addInput = this.addInput.bind(this);
    this.addSample = this.addSample.bind(this);
    this.addFilter = this.addFilter.bind(this);
    this.addProjection = this.addProjection.bind(this);
    this.addJoin = this.addJoin.bind(this);
    this.addGroup = this.addGroup.bind(this);
    this.setOutputMeta = this.setOutputMeta.bind(this);
    this.handleSave = this.handleSave.bind(this);
    this.handleCheck = this.handleCheck.bind(this);
    this.handleOutputChange = this.handleOutputChange.bind(this);
  }
  async componentDidMount() {
    const { name, path } = this.props.location.query;
    this.props.loadTableMetadata(path);
    try {
      const pipeline = await getPipeline(path, name);
      this.setState({ pipeline });
    } catch (msg) {
      console.log("getPipeline", msg);
      alert(msg.data);
    }
  }
  addInput() {
    let c = {
      name: "新输入表",
      input: {
        metaName: ""
      }
    };
    this.state.pipeline.collections.push(c);
    this.setState({ editingCollection: c });
  }
  addSample() {
    let c = {
      name: "新取样表",
      sample: {
        input: null,
        rate: 1
      }
    };
    this.state.pipeline.collections.push(c);
    this.setState({ editingCollection: c });
  }
  addFilter() {
    let c = {
      name: "新过滤表",
      filter: {
        input: null,
        expression: null
      }
    };
    this.state.pipeline.collections.push(c);
    this.setState({ editingCollection: c });
  }
  addProjection() {
    let c = {
      name: "新计算表",
      projection: {
        input: null,
        fields: []
      }
    };
    this.state.pipeline.collections.push(c);
    this.setState({ editingCollection: c });
  }
  addJoin() {
    let c = {
      name: "新连接表",
      join: {
        leftInput: null,
        rightInput: null,
        conditions: [],
        leftFields: [],
        rightFields: [],
        method: "FULL"
      }
    };
    this.state.pipeline.collections.push(c);
    this.setState({ editingCollection: c });
  }
  addGroup() {
    let c = {
      name: "新分组表",
      group: {
        input: null,
        keys: [],
        fields: []
      }
    };
    this.state.pipeline.collections.push(c);
    this.setState({ editingCollection: c });
  }
  setOutputMeta() {
    this.setState({ editingCollection: null });
  }
  removeCollection(index) {
    this.state.pipeline.collections.splice(index, 1);
    this.forceUpdate();
  }
  moveUpCollection(index) {
    if (index > 0) {
      let { collections } = this.state.pipeline;
      let prev = collections[index - 1];
      collections[index - 1] = collections[index];
      collections[index] = prev;
      this.forceUpdate();
    }
  }
  async handleSave(type) {
    let pipeline;
    try {
      pipeline = pipelineToObj(this.state.pipeline);
      pipeline.tax_type = this.refs.tax_type.value;
      pipeline.industry_code = this.refs.industry_code.value;
      pipeline.industry_name = this.refs.industry_name.value;
    } catch (msg) {
      alert("公式错误：" + msg);
      return;
    }
    try {
      await updatePipeline(pipeline);
      alert("保存成功");
    } catch (msg) {
      alert("出错了" + msg.data);
    }
  }
  async handleCheck() {
    const { path, name } = this.state.pipeline;
    try {
      await checkPipeline(path, name);
      alert("公式正确");
    } catch (msg) {
      console.log(msg);
      alert("公式错误:" + msg.obj.error);
    }
  }

  addTag = () => {
    const key = this.refs.new_tag_key.value;
    const val = this.refs.new_tag_value.value;
    this.state.pipeline.tags[key] = val;
    this.forceUpdate();
  };

  deleteTag = key => {
    delete this.state.pipeline.tags[key];
    this.forceUpdate();
  };
  handleOutputChange(output) {
    let { pipeline } = this.state;
    pipeline.resultMetaName = output;
    this.setState({ pipeline: pipeline });
  }
  render() {
    if (this.state.pipeline == null) return <h1>正在加载，请稍候</h1>;
    const { collections, tags } = this.state.pipeline;
    const { editingCollection, showingFieldOf, pipeline } = this.state;
    const { name, path } = this.props.location.query;
    return (
      <div className={styles.container}>
        <div className={styles.left}>
          <a href="javascript:" onClick={() => hashHistory.push("/")}>
            返回主页
          </a>
          <br />
          <div>
            {name} - {path}
          </div>
          {/*<a href="javascript:" className={styles.op} onClick={this.addSample}>加取样表</a>*/}
          <a href="javascript:" className={styles.op} onClick={this.addInput}>
            加输入表
          </a>
          <a href="javascript:" className={styles.op} onClick={this.addFilter}>
            加过滤表
          </a>
          <a
            href="javascript:"
            className={styles.op}
            onClick={this.addProjection}
          >
            加计算表
          </a>
          <br />
          <a href="javascript:" className={styles.op} onClick={this.addJoin}>
            加连接表
          </a>
          <a href="javascript:" className={styles.op} onClick={this.addGroup}>
            加分组表
          </a>
          <a
            href="javascript:"
            className={styles.op}
            onClick={this.setOutputMeta}
          >
            输出模板
          </a>
          <div className={styles.collections}>
            {collections.map((collection, index) => (
              <div key={index} className={styles.collection}>
                <Collection collection={collection} />
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={() =>
                    this.setState({ editingCollection: collection })}
                >
                  改
                </a>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={() => this.setState({ showingFieldOf: collection })}
                >
                  列
                </a>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.moveUpCollection.bind(this, index)}
                >
                  上
                </a>
                <a
                  href="javascript:"
                  className={styles.op}
                  onClick={this.removeCollection.bind(this, index)}
                >
                  删
                </a>
              </div>
            ))}
          </div>
          税种：<input
            type="text"
            ref="tax_type"
            placeholder="税种"
            defaultValue={pipeline.taxType}
          />
          <br />
          行业代码：<input
            type="text"
            ref="industry_code"
            placeholder="行业代码"
            defaultValue={pipeline.industryCode}
          />
          <br />
          行业名称：<input
            type="text"
            ref="industry_name"
            placeholder="行业名称"
            defaultValue={pipeline.industryName}
          />
          <br />
          已有标签:
          {Object.keys(tags).map(tag => (
            <div style={{ display: "flex", flexDirection: "row" }}>
              <button type="button" onClick={() => this.deleteTag(tag)}>
                删除
              </button>
              <div>{tag}</div>:
              <div>{tags[tag]}</div>
            </div>
          ))}
          <br />
          名称 <input type="text" ref="new_tag_key" />
          <br />
          值 <br />{" "}
          <textarea style={{ height: 200 }} type="text" ref="new_tag_value" />
          <button type="button" onClick={this.addTag}>
            添加新标签
          </button>
          <br />
          <button type="button" onClick={this.handleSave}>
            保存
          </button>
          <button type="button" onClick={this.handleCheck}>
            检查
          </button>
          <br />
          <h4>公式模板</h4>
          <ul>
            <li>
              <Expression expression={{ literal: { booleanValue: true } }} />
            </li>
            <li>
              <Expression expression={{ literal: { booleanValue: false } }} />
            </li>
            <li>
              <Expression
                expression={{ literal: { floatValue: 0.0 } }}
                readOnly
              />
            </li>
            <li>
              <Expression
                expression={{ literal: { stringValue: "" } }}
                readOnly
              />
            </li>
            <li>
              <Expression
                expression={{ operation: { operator: "ADD", operands: [] } }}
                readOnly
              />
            </li>
          </ul>
        </div>
        <div className={styles.right}>
          <div className={styles.fieldList}>
            {showingFieldOf && (
              <FieldList path={path} collection={showingFieldOf} />
            )}
          </div>
          <div className={styles.content}>
            <CollectionEditor
              path={path}
              collection={editingCollection}
              output={pipeline.resultMetaName}
              handleOutputChange={this.handleOutputChange}
            />
          </div>
        </div>
      </div>
    );
  }
}

export default connect(null, { loadTableMetadata })(
  DragDropContext(HTML5Backend)(PipelineEdit)
);
