import React from 'react';
import {DragSource} from 'react-dnd';
import {expressionType, literalType, OPERATOR_TEXT} from './types';
import Field from './field';
import ExpressionHolder from './expression-holder';
import styles from './expression.css';

const expressionSource = {
  beginDrag(props) {
    return JSON.parse(JSON.stringify(props.expression));
  }
}

function collectDrag(connect, monitor) {
  return {
    connectDragSource: connect.dragSource(),
    isDragging: monitor.isDragging()
  }
}

class Expression extends React.Component {
  constructor(props) {
    super(props);
    this.state = {collapsed: false, hover: false};
    this.handleFloatValueChange = this.handleFloatValueChange.bind(this);
    this.handleStringValueChange = this.handleStringValueChange.bind(this);
    this.handleOperatorChange = this.handleOperatorChange.bind(this);
    this.handleDropOperand = this.handleDropOperand.bind(this);
    this.handleMouseOver = this.handleMouseOver.bind(this);
    this.handleMouseOut = this.handleMouseOut.bind(this);
  }
  handleMouseOver(e) {
    this.setState({hover: true});
    e.stopPropagation();
  }
  handleMouseOut(e) {
    this.setState({hover: false});
    e.stopPropagation();
  }
  handleFloatValueChange(e) {
    this.props.expression.literal.floatValue = Number(e.target.value);
    this.forceUpdate();
  }
  handleStringValueChange(e) {
    this.props.expression.literal.stringValue = e.target.value;
    this.forceUpdate();
  }
  handleOperatorChange(e) {
    if (this.props.readOnly) {
      e.preventDefault();
      return;
    }
    this.props.expression.operation.operator = e.target.value;
    this.forceUpdate();
  }
  handleDropOperand(operand) {
    if (this.props.readOnly) return;
    this.props.expression.operation.operands.push(operand);
    this.forceUpdate();
  }
  removeOperand(index) {
    this.props.expression.operation.operands.splice(index, 1);
    this.forceUpdate();
  }
  moveOperandUp(index) {
    if (index == 0) return;
    let {operands} = this.props.expression.operation;
    let previous = operands[index - 1];
    operands[index - 1] = operands[index];
    operands[index] = previous;
    this.forceUpdate();
  }
  render() {
    const {expression, connectDragSource, isDragging, readOnly} = this.props;
    const {collapsed, hover} = this.state;
    const type = expressionType(expression);
    const className = hover? styles.expressionHover: styles.expression;
    if (type == 'FIELD') {
      return <Field field={expression.field} />;
    } else if (type == 'LITERAL') {
      const {literal} = expression;
      const lType = literalType(literal);
      if (lType == 'BOOLEAN')
        return connectDragSource(
          <div className={className} onMouseOver={this.handleMouseOver} onMouseOut={this.handleMouseOut}>
            <span className={styles.label}>{literal.booleanValue? '真': '假'}</span>
          </div>
        );
      if (lType == 'FLOAT')
        return connectDragSource(
          <div className={className} onMouseOver={this.handleMouseOver} onMouseOut={this.handleMouseOut}>
            <span className={styles.label}>实数</span>
            <input type="number" step="0.0001" placeholder="值" value={literal.floatValue} onChange={this.handleFloatValueChange} readOnly={readOnly}/>
          </div>
        );
      if (lType == 'STRING')
        return connectDragSource(
          <div className={className} onMouseOver={this.handleMouseOver} onMouseOut={this.handleMouseOut}>
            <span className={styles.label}>字串</span>
            <input type="text" placeholder="内容" value={literal.stringValue} onChange={this.handleStringValueChange} readOnly={readOnly}/>
          </div>
        );
    } else if (type == 'OPERATION') {
      const {operation} = expression;
      return connectDragSource(
        <div className={className} onMouseOver={this.handleMouseOver} onMouseOut={this.handleMouseOut} style={{visibility: isDragging?'hidden':'inherit'}}>
          {collapsed || <a href="javascript:" className={styles.op} title="折叠" onClick={()=>this.setState({collapsed: true})}>▾</a>}
          {collapsed && <a href="javascript:" className={styles.op} title="展开" onClick={()=>this.setState({collapsed: false})}>▴</a>}
          <span className={styles.label}>运算</span>
          <select className={styles.operator} value={operation.operator} onChange={this.handleOperatorChange}>
            {Object.keys(OPERATOR_TEXT).map(operator =>
              <option key={operator} value={operator}>{OPERATOR_TEXT[operator]}</option>
            )}
          </select>
          {collapsed || <div className={styles.operands}>
            {operation.operands.map((operand, index) =>
              <div className={styles.operand} key={index}>
                <a href="javascript:" className={styles.op} title="上移" onClick={this.moveOperandUp.bind(this, index)}>⇪</a>
                <a href="javascript:" className={styles.op} title="删除" onClick={this.removeOperand.bind(this, index)}>×</a>
                <DndExpression expression={operand}/>
              </div>
            )}
            <ExpressionHolder onChange={this.handleDropOperand}/>
          </div>}
        </div>
      );
    }
  }
}

const DndExpression = DragSource('EXPRESSION', expressionSource, collectDrag)(Expression);
//const DndExpression = DropTarget(['EXPRESSION', 'FIELD'], expressionSource, collectDrop)(DragExpression);
export default DndExpression;