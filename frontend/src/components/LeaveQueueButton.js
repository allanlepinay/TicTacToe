import React from 'react';
import axios from 'axios';

const LeaveQueueButton = () => {
  const handleLeaveQueue = async () => {
      const username = localStorage.getItem('username');
      if (!username || username == "") {
        console.error('Username not found in local storage');
        return;
      }
      // Todo, probable probleme de synchronisite
      // todo add token refresh if clicked ?
      await axios.post('/leave-queue', {           
        params: {
        // todo this have to be changed when user check will be refractor
        username: username
      }
    })
    .then(response => {
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
