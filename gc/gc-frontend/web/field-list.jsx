import React from 'react';
import Field from './field';
import {collectionFields} from './types';
import styles from './field-list.css';

class FieldList extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      filter: ''
    };
    this.handleFilterChange = this.handleFilterChange.bind(this);
  }
  handleFilterChange(e) {
    this.setState({filter: e.target.value});
  }
  render() {
    const {collection, path} = this.props;
    const {filter} = this.state;
    let fields = collectionFields(collection, path);
    if (fields == null) return <span className={styles.error}>该表存在循环引用，无法列出其中的列！</span>;
    if (filter.trim().length > 0) {
      fields = fields.filter(field => field.name.toLowerCase().indexOf(filter.toLowerCase()) != -1);
    }
    fields.sort((fa, fb) => {
      if (fa.name < fb.name) return -1;
      if (fa.name == fb.name) return 0;
      return 1;
    });
    return (
      <div>
        <h2>{collection.name}</h2>
        <input type="text" size="100" onChange={this.handleFilterChange} value={filter} placeholder="过滤，留空则显示全部" /><br/>
        {fields.map((field, index) =>
          <div key={index} className={styles.field}>
            <Field field={field} />
          </div>
        )}
      </div>
    );
  }
}
export default FieldList;