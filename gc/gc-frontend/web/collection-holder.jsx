import React from 'react';
import {DropTarget} from 'react-dnd';
import Collection from './collection';
import styles from './collection-holder.css';

const collectionTarget = {
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

class CollectionHolder extends React.Component {
  render() {
    const {connectDropTarget, isOver, collection} = this.props;
    return connectDropTarget(
      <div className={isOver?styles.holderActive:styles.holderNormal}>
        {collection && <Collection collection={collection} />}
      </div>
    );
  }
}

export default DropTarget('COLLECTION', collectionTarget, collect)(CollectionHolder);