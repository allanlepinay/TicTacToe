import React from 'react';
import '../axiosConfig';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';


const LeaveQueueButton = () => {
  const navigate = useNavigate();

  const handleLeaveQueue = () => {
      const username = localStorage.getItem('username');
      if (!username || username == "") {
        console.error('Username not found in local storage');
        return;
      }
      // todo add token refresh if clicked ?
      axios.post('/leave-queue', {
        username: localStorage.getItem('username')
      })
      .then(response => {
        navigate("/");
        window.location.reload();
        console.log('Successfully left the queue:', response.data);
      })
      .catch(error => {
        console.error('Error leaving the queue:', error);
      });        
  };

  return (
    <button onClick={handleLeaveQueue}>
      Leave Queue
    </button>
  );
};

export default LeaveQueueButton;
