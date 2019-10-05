import {combineReducers} from 'redux';

function tableMetadata(state={}, action) {
  switch(action.type) {
  case 'LOAD_TABLE_METADATA_SUCCESS':
    let metas = {};
    for (let meta of action.metadata) {
      metas[meta.name] = meta;
    }
    return {
      ...state,
      [action.path]: metas
    };
  default:
    return state;
  }
}

function databases(state={}, action) {
  switch(action.type) {
  case 'LOAD_DATABASES_SUCCESS':
    return {
      ...state,
      [action.path]: {
        ...state[action.path],
        [action.name]: action.databases
      }
    };
  default:
    return state;
  }
}

export default combineReducers({
  tableMetadata,
  databases
});