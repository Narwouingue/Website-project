
// Make an HTTP request to the server to get the token
var token
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
    token = data.token;
    // Use the token as needed
    console.log('Token:', token);
    // Further processing of the token
  })
  .catch(error => {
    // Handle any errors that occurred during the fetch operation
    console.error('Error:', error);
  });
// option 1, set room defaults
const room = new Room({
  audioCaptureDefaults: {
    autoGainControl: true,
    deviceId: '',
    echoCancellation: true,
    noiseSuppression: true,
  },
  videoCaptureDefaults: {
    deviceId: '',
    facingMode: 'user',
    resolution: {
      width: 1920,
      height: 1080,
      frameRate: 30,
    },
  },
  publishDefaults: {
    videoEncoding: {
      maxBitrate: 1_500_000,
      maxFramerate: 30,
    },
    screenShareEncoding: {
      maxBitrate: 1_500_000,
      maxFramerate: 30,
    },
    audioBitrate: 20_000,
    dtx: true,
    // only needed if overriding defaults
  },
})



import { Room } from 'livekit-client';

const wsURL = "wss://website-4swwhc1o.livekit.cloud"


await room.connect(wsURL, token);
console.log('connected to room', room.name);

// publish local camera and mic tracks
await room.localParticipant.enableCameraAndMicrophone();

localParticipant.setTrackSubscriptionPermissions(false, [
  {
    participantIdentity: "allowed-identity",
    allowAll: true,
  }
])

import {
  connect,
  RoomEvent,
} from 'livekit-client';

room
  .on(RoomEvent.TrackSubscribed, handleTrackSubscribed)

/*function handleTrackSubscribed(
  track: RemoteTrack,
  publication: RemoteTrackPublication,
  participant: RemoteParticipant
) {
  
}*/

const element = track.attach();
parentElement.appendChild(element);