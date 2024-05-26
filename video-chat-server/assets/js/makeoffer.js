const servers = {
  iceServers: [
    {
      urls: ['stun:stun1.l.google.com:19302', 'stun:stun2.l.google.com:19302']
    },
  ],
  iceCandidiatePoolSize: 10,
};
let conn;
let peerConn = new RTCPeerConnection(servers);
let localStream = null;
let remoteStream = null;

const webcamButton = document.getElementById("enable-webcam");
const webcamVideo = document.getElementById("local-video");
const remoteVideo = document.getElementById("remote-video");


function submitOfferName() {
  let inputButton = document.getElementById("user-input");
  let inputValue = inputButton.value;
  let hash = "";

  fetch("http://127.0.0.1:4009/videocall/OfferValidation", {
    method: "POST",
    body: inputValue,
  })
  .then((response) => response.text())
  .then((text) => {
    if (text === "NO") {
        console.log("it said no")
        return
        // have a popup say that name is taken
      } else {
        console.log(text);
        inputButton.disabled = true;
        hash = text;

        if (window['WebSocket']) {
          conn = new WebSocket('ws://' + document.location.host + '/videocall/MakeOffer/ws');
          //let initial = true;
          let answerList = document.getElementById("answer-list")




          conn.onopen = function () {
            conn.send(inputValue + " " + hash)
          }

          conn.onclose = () => {
            console.log("Connection closing")
          }


          conn.onmessage = evt => {
            console.log("MESSAGE RECIEVED: " + evt.data)

            //if (initial) {
            //  console.log(evt.data)
            //  initial = false
            //}

            if (evt.data.slice(0,3) === "<li") {
              console.log("Made it here")
              // add it to the list
              answerList.innerHTML = evt.data
            } else if (evt.data === "DONE") {
              console.log("wooo bye")
              conn.send("DONE")
              conn.close()
            } else {
              console.log(evt.data)
            }

            
          }


        }
      }

    });

  // once the offer and answers have been exchanged we should end the websocket connection and just be in the call.
  //


}

async function clickName(name) {
  /*
  const offerDescription = await peerConn.createOffer();
  await peerConn.setLocalDescription(offerDescription);
  
  const offer = {
    sdp: offerDescription.sdp,
    type: offerDescription.type,
  };


  console.log(offer)
  */
  let offer = "meme"
  if (conn !== null) {
    conn.send("request{ name: " + name + " sdp: " + offer);
  } 
} 

webcamButton.onclick = async () => {
  // you need to remove the audio track from the local stream because it causes an echo
  localStream = await navigator.mediaDevices.getUserMedia({video: true, audio: true});
  remoteStream = new MediaStream();

  localStream.getTracks().forEach((track) => {
    peerConn.addTrack(track, localStream);
  });

  peerConn.ontrack = event => {
    event.streams[0].getTracks().forEach(track => {
      remoteStream.addTrack(track);
    });
  };

  webcamVideo.srcObject = localStream;
  remoteVideo.srcObject = remoteStream;

};

