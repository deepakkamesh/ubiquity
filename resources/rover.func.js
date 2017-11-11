// Constants.
// Handle websocket registrations and update Rover data panels.

var CmdType = {
    ERR: 0,
    CMD: 1,
    AUDIO_START: 2,
    AUDIO_STOP: 3,
    DRIVE_FWD: 4,
    DRIVE_BWD: 5,
    DRIVE_LEFT: 6,
    DRIVE_RIGHT: 7,
    SERVO_UP: 8,
    SERVO_DOWN: 9,
    SERVO_STEP: 10,
}

// Control message handlers
$(document).ready(function() {
    var errorContainer = document.querySelector('#error-popup');

    wsCtrl = new WebSocket("wss://" + window.location.host + "/control");
    wsCtrl.onopen = function(evt) {
        $("#conn_spinner").show();
    }

    wsCtrl.onclose = function(evt) {
        $("#conn_spinner").hide();
        wsCtrl = null;
    }

    wsCtrl.onmessage = function(evt) {
        msg = JSON.parse(evt.data);
        console.log(msg);

        switch (msg.CmdType) {
            case CmdType.ERR:
                var err = {
                    message: 'Error: ' + msg.Data
                };
                errorContainer.MaterialSnackbar.showSnackbar(err);
        }
    }

    wsCtrl.onerror = function(evt) {
        print("Control ERROR: " + evt.data);
    }
});

// Callback for keyboard keys Drive Control.
$(document).keydown(function(e) {
    var cmd
    switch (e.which) {
        case 37:
            cmd = CmdType.DRIVE_LEFT;
            break;
        case 38:
            cmd = CmdType.DRIVE_FWD;
            break;
        case 39:
            cmd = CmdType.DRIVE_RIGHT;
            break;
        case 40:
            cmd = CmdType.DRIVE_BWD;
            break;
        default:
            cmd = -1;
    }
    if (cmd == -1) {
        return
    }
    var cmd = JSON.stringify({
        CmdType: cmd,
        Data: parseInt($('#drive_velocity_sel').val()),
    });
    console.log(cmd);

    wsCtrl.send(cmd);
});

// Servo and Drive Controls.
$(document).ready(function() {
    // Drive velocity selector.
    var driveVelSel = document.querySelector('#drive_velocity_sel');
    driveVelSel.addEventListener('click', function() {
        val = $('#drive_velocity_sel').val();
        $("#drive_velocity_sel_disp").empty()
        $("#drive_velocity_sel_disp").append(val);
    });

    // Servo Controls.
    var servoDownButton = document.querySelector('#servo-down');
    servoDownButton.addEventListener('click', function() {
        var cmd = JSON.stringify({
            CmdType: CmdType.SERVO_DOWN,
        });
        console.log(cmd);
        wsCtrl.send(cmd);
    });

    var servoUpButton = document.querySelector('#servo-up');
    servoUpButton.addEventListener('click', function() {
        var cmd = JSON.stringify({
            CmdType: CmdType.SERVO_UP,
        });
        console.log(cmd);
        wsCtrl.send(cmd);
    });

    // TODO: Change to center and max or something.
    var servoLeftButton = document.querySelector('#servo-left');
    servoLeftButton.addEventListener('click', function() {});

    var servoRightButton = document.querySelector('#servo-right');
    servoRightButton.addEventListener('click', function() {});

    // Set the step for Servo.
    var servoAngleDeltaSel = document.querySelector('#servo_angle_step');
    servoAngleDeltaSel.addEventListener('click', function() {
        val = $('#servo_angle_step').val();
        $("#servo_angle_step_disp").empty();
        $("#servo_angle_step_disp").append(val);

        var cmd = JSON.stringify({
            CmdType: CmdType.SERVO_STEP,
            Data: parseInt(val),
        });
        console.log(cmd);
        wsCtrl.send(cmd);
    });
});

// Audio handlers.
$(document).ready(function() {

    ws = new WebSocket("wss://" + window.location.host + "/audiostream");

    ws.onerror = function(evt) {
        print("ERROR: " + evt.data);
    }

    // downsampleBuffer downsamples and converts to uint16.
    var downsampleBuffer = function(buffer, sampleRate, outSampleRate) {
        if (outSampleRate == sampleRate) {
            return buffer;
        }
        if (outSampleRate > sampleRate) {
            throw "downsampling rate show be smaller than original sample rate";
        }
        var sampleRateRatio = sampleRate / outSampleRate;
        var newLength = Math.round(buffer.length / sampleRateRatio);
        var result = new Int16Array(newLength);
        var offsetResult = 0;
        var offsetBuffer = 0;
        while (offsetResult < result.length) {
            var nextOffsetBuffer = Math.round((offsetResult + 1) * sampleRateRatio);
            var accum = 0,
                count = 0;
            for (var i = offsetBuffer; i < nextOffsetBuffer && i < buffer.length; i++) {
                accum += buffer[i];
                count++;
            }

            result[offsetResult] = Math.min(1, accum / count) * 0x7FFF;
            offsetResult++;
            offsetBuffer = nextOffsetBuffer;
        }
        return result;
    }

    /*
        function floatTo16Bit(inputArray, startIndex) {
            var output = new Uint16Array(inputArray.length - startIndex);
            for (var i = 0; i < inputArray.length; i++) {
                var s = Math.max(-1, Math.min(1, inputArray[i]));
                output[i] = s < 0 ? s * 0x80 : s * 0x7F;
            }
            return output;
        } */

    // Audio stream handling.
    var streamControl;
    var handleSuccess = function(stream) {
        var context = new AudioContext();
        var source = context.createMediaStreamSource(stream);
        var processor = context.createScriptProcessor(8192, 1, 1);

        source.connect(processor);
        processor.connect(context.destination);

        processor.onaudioprocess = function(e) {
            if (!streamControl) {
                return;
            }

            var ib = e.inputBuffer;
            var i = ib.getChannelData(0);
            //var conv = floatTo16Bit(i, 0);
            var conv = downsampleBuffer(i, 44100, 4000);
            // console.log(conv)
            ws.send(conv);
        };
    };
    navigator.mediaDevices.getUserMedia({
            audio: true,
            video: false
        })
        .then(handleSuccess);



    var recordAudioBtn = document.querySelector('#record_audio');
    recordAudioBtn.addEventListener('mousedown', function() {
        streamControl = true;
        console.log("started");
    });
    recordAudioBtn.addEventListener('touchstart', function() {
        streamControl = true;
        console.log("started");
    });

    recordAudioBtn.addEventListener('mouseup', function() {
        streamControl = false;
        console.log("stop");
    });
    recordAudioBtn.addEventListener('touchend', function() {
        streamControl = false;
        console.log("stop");
    });

});
