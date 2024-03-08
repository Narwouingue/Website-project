
// Make an HTTP request to the server to get the token
fetch('/getToken')
  .then(response => {
    // Check if the response is successful
    if (!response.ok) {
      throw new Error('Failed to fetch token');
    }
    // Parse the JSON response
    return response.json();
  })
  .then(data => {
    // Retrieve the token from the response data
    const token = data.token;
    // Use the token as needed
    console.log('Token:', token);
    // Further processing of the token
  })
  .catch(error => {
    // Handle any errors that occurred during the fetch operation
    console.error('Error:', error);
  });



import { Room } from 'livekit-client';

const wsURL = "wss://website-4swwhc1o.livekit.cloud"

const room = new Room();
await room.connect(wsURL, token);
console.log('connected to room', room.name);

// publish local camera and mic tracks
await room.localParticipant.enableCameraAndMicrophone();