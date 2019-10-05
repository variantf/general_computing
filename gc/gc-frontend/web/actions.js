import * as webapi from './webapi';

export function loadTableMetadata(path) {
  return async dispatch => {
    try {
      dispatch({
        type: 'LOAD_TABLE_METADATA_SUCCESS',
        path,
        metadata: await webapi.listTableMetadata(path, false)
      });
    } catch (msg) {
      console.log(msg);
    }
  };
}

export function loadDatabases(path, name) {
  return async dispatch => {
    try {
      dispatch({
        type: 'LOAD_DATABASES_SUCCESS',
        path,
        name,
        databases: await webapi.listDatabaseTable(path, name)
      });
    } catch (msg) {
      console.log(msg);
    }
  };
}