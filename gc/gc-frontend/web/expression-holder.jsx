import React from 'react';
import {DropTarget} from 'react-dnd';
import Expression from './expression';
import styles from './expression-holder.css';

const expressionTarget = {
  drop(props, monitor) {
    if (!monitor.isOver({shallow: true})) {
      return;
    }
    if (props.needConfirm && !confirm('确认全部替换吗?')) return;
    if (monitor.getItemType() == 'FIELD') {
      props.onChange({
        field: monitor.getItem()
      });
    } else {
      props.onChange(monitor.getItem());
    }
  }
}

function collect(connect, monitor) {
  return {
    connectDropTarget: connect.dropTarget(),
    isOver: monitor.isOver({shallow: true})
  }
}

class ExpressionHolder extends React.Component {
  render() {
    const {connectDropTarget, isOver, expression} = this.props;
    return connectDropTarget(
      <div className={isOver?styles.holderActive:styles.holderNormal}>
        {expression && <Expression expression={expression} />}
      </div>
    );
  }
}

export default DropTarget(['FIELD','EXPRESSION'], expressionTarget, collect)(ExpressionHolder);