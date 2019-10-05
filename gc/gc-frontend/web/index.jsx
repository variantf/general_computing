import React from 'react';
import ReactDOM from 'react-dom';
import {Provider} from 'react-redux';
import {Router, Route, hashHistory} from 'react-router';
import store from './store';
import DevTools from './dev-tools';
import PipelineList from './pipeline-list';
import PipelineEdit from './pipeline-edit';
import TaskEdit from './task-edit';
import TableMetadataList from './table-metadata-list'
import DatabaseTableList from './database-table-list'
import UploadPage from './upload'
// import Login from './Login'

ReactDOM.render(
  <Provider store={store}>
    <div>
      <DevTools />
      <Router history={hashHistory}>
        <Route path="/" component={PipelineList}/>
        <Route path="/pipeline" component={PipelineEdit} />
        <Route path="/task" component={TaskEdit} />
        <Route path="/tablemetadata" component={TableMetadataList} />
        <Route path="/databasetable" component={DatabaseTableList} />
        <Route path="/upload" component={UploadPage} />
      </Router>
    </div>
  </Provider>,
  document.getElementById('app')
);