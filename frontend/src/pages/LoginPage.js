import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import '../axiosConfig';
import axios from 'axios';

function LoginPage() {
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const handleSubmit = (e) => {
    e.preventDefault();

    axios.post('/login', { name: name, password: password })
    .then(response => {
      console.log('Login successful:', response.data);
      const { access_token, refresh_token } = response.data;
      localStorage.setItem('token', access_token);
      localStorage.setItem('refresh_token', refresh_token);
      // todo probably not the best thing to send to check who is the user
      localStorage.setItem('username', name);
      navigate("/lobby");
      window.location.reload();
    })
    .catch(error => {
        console.error('Error login:', error);
    });        
  };

  return (
    <div>
      <h1>Login</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label>Name:</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Password:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button type="submit">Login</button>
      </form>
    </div>
  );
}

export default LoginPage;