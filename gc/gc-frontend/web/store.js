import {createStore, compose, applyMiddleware} from 'redux';
import reduxThunk from 'redux-thunk';
import reduxPromise from 'redux-promise';
import DevTools from './dev-tools';
import reducer from './reducer';

const enhancer = compose(
  applyMiddleware(reduxThunk, reduxPromise),
  DevTools.instrument()
);

export default createStore(reducer, undefined, enhancer);
