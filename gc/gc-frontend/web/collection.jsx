import React from 'react';
import {DragSource} from 'react-dnd';
import {COLLECTION_TYPE_TEXT, collectionType} from './types';
import styles from './collection.css';

const collectionSource = {
  beginDrag(props) {
    return props.collection;
  }
}

function collect(connect, monitor) {
  return {
    connectDragSource: connect.dragSource()
  }
}

class Collection extends React.Component {
  render() {
    const {connectDragSource, collection} = this.props;
    return connectDragSource(
      <span className={styles.collection}>
        {collection.name}
        <span className={styles.type}>{COLLECTION_TYPE_TEXT[collectionType(collection)]}</span>
      </span>
    );
  }
}

export default DragSource('COLLECTION', collectionSource, collect)(Collection);