import { configureStore } from '@reduxjs/toolkit';
import websocketReducer from './websocketSlice';

const store = configureStore({
  reducer: {
    websocket: websocketReducer,
  },
});

export default store;