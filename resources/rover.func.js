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
    VIDEO_ENABLE: 11,
    VIDEO_DISABLE: 12,
    AUDIO_ENABLE: 13,
    AUDIO_DISABLE: 14,
    MASTER_ENABLE: 15,
    MASTER_DISABLE: 16,
    SERVO_ABS: 17,
    DRIVE_LEFT_ONLY: 18,
    DRIVE_RIGHT_ONLY: 19,
    HEADLIGHT_ON: 20,
    HEADLIGHT_OFF: 21,
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

    SendControlCmd = function(cmd, data) {
        cmdJS = JSON.stringify({
            CmdType: cmd,
            Data: data,
        });
        console.log(cmdJS);
        wsCtrl.send(cmdJS);
    }
});

// Callback for keyboard keys Drive Control.
$(document).keydown(function(e) {
    var cmd
    switch (e.which) {
        case 37:
            if (document.getElementById('rotate_dual').checked) {
                cmd = CmdType.DRIVE_LEFT;
                break;
            }
            cmd = CmdType.DRIVE_LEFT_ONLY;
            break;
        case 38:
            cmd = CmdType.DRIVE_FWD;
            break;
        case 39:
            if (document.getElementById('rotate_dual').checked) {
                cmd = CmdType.DRIVE_RIGHT;
                break;
            }
            cmd = CmdType.DRIVE_RIGHT_ONLY;
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
    SendControlCmd(cmd, parseInt($('#drive_velocity_sel').val()));
});

// Control
$(document).ready(function() {
    var masterEnable = document.querySelector('#master_enable');
    masterEnable.addEventListener('click', function() {
        if (document.getElementById('master_enable').checked) {
            SendControlCmd(CmdType.MASTER_ENABLE);
        } else {
            SendControlCmd(CmdType.MASTER_DISABLE);
        }
    });

    var audioEnable = document.querySelector('#audio_enable');
    audioEnable.addEventListener('click', function() {
        if (document.getElementById('audio_enable').checked) {
            SendControlCmd(CmdType.AUDIO_ENABLE);
        } else {
            SendControlCmd(CmdType.AUDIO_DISABLE);
        }
    });

    document.querySelector('#headlight_enable').addEventListener('click', function() {
        if (document.getElementById('headlight_enable').checked) {
            SendControlCmd(CmdType.HEADLIGHT_ON);
        } else {
            SendControlCmd(CmdType.HEADLIGHT_OFF);
        }
    });



    var videoEnable = document.querySelector('#video_enable');
    videoEnable.addEventListener('click', function() {
        fps = parseInt($('#fps_sel').val());
        resMode = parseInt($('#res-sel').val());
        data = [fps, resMode];
        if (document.getElementById('video_enable').checked) {
            SendControlCmd(CmdType.VIDEO_ENABLE, data);
            $("#video_stream").attr("src", "/videostream" + '?' + Math.random());
        } else {
            $("#video_stream").attr("src", "");
            SendControlCmd(CmdType.VIDEO_DISABLE, data);
        }
    });
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
    document.querySelector('#servo-down').addEventListener('click', function() {
        SendControlCmd(CmdType.SERVO_DOWN);
    });

    document.querySelector('#servo-up').addEventListener('click', function() {
        SendControlCmd(CmdType.SERVO_UP);
    });

    document.querySelector('#servo-top').addEventListener('click', function() {
        SendControlCmd(CmdType.SERVO_ABS, 0);
    });

    document.querySelector('#servo-center').addEventListener('click', function() {
        SendControlCmd(CmdType.SERVO_ABS, 90);
    });

    document.querySelector('#servo-bottom').addEventListener('click', function() {
        SendControlCmd(CmdType.SERVO_ABS, 180);
    });

    // Set the step for Servo.
    var servoAngleDeltaSel = document.querySelector('#servo_angle_step');
    servoAngleDeltaSel.addEventListener('click', function() {
        val = $('#servo_angle_step').val();
        $("#servo_angle_step_disp").empty();
        $("#servo_angle_step_disp").append(val);

        SendControlCmd(CmdType.SERVO_STEP, parseInt(val));
    });
});

// Audio handlers.
$(document).ready(function() {

    ws = new WebSocket("wss://" + window.location.host + "/audiostream");
    ws.binaryType = 'arraybuffer';

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

    function int16ToFloat32(inputArray, startIndex, length) {
        var output = new Float32Array(inputArray.length - startIndex);
        for (var i = startIndex; i < length; i++) {
            var int = inputArray[i];
            // If the high bit is on, then it is a negative number, and actually counts backwards.
            var float = (int >= 0x8000) ? -(0x10000 - int) / 0x8000 : int / 0x7FFF;
            output[i] = float;
        }
        return output;
    }

    // Send audio packets from browser to Ubiquity.
    var streamControl;
    var handleSuccess = function(stream) {
        var context = new AudioContext();
        var source = context.createMediaStreamSource(stream);
        var processor = context.createScriptProcessor(2048, 1, 1);

        source.connect(processor);
        processor.connect(context.destination);

        processor.onaudioprocess = function(e) {
            if (!streamControl) {
                return;
            }
            var ib = e.inputBuffer;
            var i = ib.getChannelData(0);
            var conv = downsampleBuffer(i, 44100, 4000);
            //console.log(conv)
            ws.send(conv);
        };
    };
    navigator.mediaDevices.getUserMedia({
            audio: true,
            video: false
        })
        .then(handleSuccess);

    // Recieve and play audio packets from Ubiquity.
    var context = new window.AudioContext()
    var channels = 1
    var sampleRate = 8000
    var frames = 1024
    var buffer = context.createBuffer(channels, frames, sampleRate)

    ws.onmessage = function(evt) {
        var data = new Int16Array(evt.data);
        var floatData = int16ToFloat32(data, 0, data.length)
        buffer.getChannelData(0).set(floatData)

        var source = context.createBufferSource()
        source.buffer = buffer
        // Then output to speaker for example
        source.connect(context.destination)
        source.start(0)
    }

    var recStart = document.querySelector('#rec-start');
    recStart.addEventListener('click', function() {
        if (document.getElementById('rec-start').checked) {
            streamControl = true;
            SendControlCmd(CmdType.AUDIO_START, '');
            console.log("Rec. started");
        } else {
            streamControl = false;
            SendControlCmd(CmdType.AUDIO_STOP, '');
            console.log("Rec. stopped");
        }
    });

});
