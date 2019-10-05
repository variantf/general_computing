import React from 'react';
import {DragSource} from 'react-dnd';
import styles from './field.css';

const fieldSource = {
  beginDrag(props) {
    return props.field;
  }
}

function collect(connect, monitor) {
  return {
    connectDragSource: connect.dragSource()
  }
}

class Field extends React.Component {
  render() {
    const {connectDragSource, field} = this.props;
    return connectDragSource(
      <span className={styles.field}>
        {field.name}
      </span>
    );
  }
}

export default DragSource('FIELD', fieldSource, collect)(Field);