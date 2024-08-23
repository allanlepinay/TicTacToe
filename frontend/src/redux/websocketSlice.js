import { createSlice } from '@reduxjs/toolkit';
const token = localStorage.getItem('token');
const websocketSlice = createSlice({
  name: 'websocket',
  initialState: {
    connection: new WebSocket(`ws://localhost:8080/ws?token=${token}`),
  },
  reducers: {
    setWebSocketConnection: (state, action) => {
      state.connection = action.payload;
    },
    closeWebSocketConnection: (state) => {
      if (state.connection) {
        state.connection.close();
        state.connection = null;
      }
    },
  },
});

export const { setWebSocketConnection, closeWebSocketConnection } = websocketSlice.actions;

export default websocketSlice.reducer;
