import React from 'react';
import {DropTarget} from 'react-dnd';
import Field from './field';
import styles from './field-holder.css';

const fieldTarget = {
  drop(props, monitor) {
    props.onChange(monitor.getItem());
  }
}

function collect(connect, monitor) {
  return {
    connectDropTarget: connect.dropTarget(),
    isOver: monitor.isOver()
  }
}

class FieldHolder extends React.Component {
  render() {
    const {connectDropTarget, isOver, field} = this.props;
    return connectDropTarget(
      <div className={isOver?styles.holderActive:styles.holderNormal}>
        {field && <Field field={field} />}
      </div>
    );
  }
}

export default DropTarget('FIELD', fieldTarget, collect)(FieldHolder);