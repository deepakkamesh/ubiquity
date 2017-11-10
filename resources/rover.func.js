// Constants.
// Handle websocket registrations and update Rover data panels.

var MsgType = {
    ERR: 0,
    CMD: 1,
    AUDIO: 2
}

var CmdType = {
    FWD: 0,
    BWD: 1,
    LEFT: 2,
    RIGHT: 3,
}

$(document).ready(function() {

    ws = new WebSocket("wss://" + window.location.host + "/datastream");

    ws.onopen = function(evt) {
        $("#conn_spinner").show();
    }

    ws.onclose = function(evt) {
        $("#conn_spinner").hide();
        ws = null;
    }

    ws.onmessage = function(evt) {
        st = JSON.parse(evt.data);
        if (st.Err != "") {
            console.log(st.Err);
        }
        piData = st.Pi;
        for (var pktID in piData) {
            pkt = piData[pktID];
            switch (pktID) {
                case "0": //I2CBus State.
                    setTimeout(function(pkt) {
                        if (pkt == 1) {
                            $('#i2c_en_label')[0].MaterialSwitch.on();
                            return;
                        }
                        $('#i2c_en_label')[0].MaterialSwitch.off();
                        return;
                    }, 100, pkt);

                case "1": // AuxPower State.
                    setTimeout(function(pkt) {
                        if (pkt == 1) {
                            $('#aux_power_label')[0].MaterialSwitch.on();
                            return;
                        }
                        $('#aux_power_label')[0].MaterialSwitch.off();
                        return;
                    }, 100, pkt);
            }
        }
    }

    ws.onerror = function(evt) {
        print("ERROR: " + evt.data);
    }

    $(document).keydown(function(e) {
        var cmd
        switch (e.which) {
            case 37:
                cmd = CmdType.LEFT;
                break;
            case 38:
                cmd = CmdType.FWD;
                break;
            case 39:
                cmd = CmdType.RIGHT;
                break;
            case 40:
                cmd = CmdType.BWD;
                break;
            default:
                cmd = -1;
        }
        if (cmd == -1) {
            return
        }
        ws.send(JSON.stringify({
            MsgType: MsgType.CMD,
            Data: {
                CmdType: cmd,
                Param: parseInt($('#drive_velocity_sel').val()),
            }
        }));
    });

    var driveVelSel = document.querySelector('#drive_velocity_sel');
    driveVelSel.addEventListener('click', function() {
        val = $('#drive_velocity_sel').val();
        $("#drive_velocity_sel_disp").empty()
        $("#drive_velocity_sel_disp").append(val);
    });

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

    function floatTo16Bit(inputArray, startIndex) {
        var output = new Uint16Array(inputArray.length - startIndex);
        for (var i = 0; i < inputArray.length; i++) {
            var s = Math.max(-1, Math.min(1, inputArray[i]));
            output[i] = s < 0 ? s * 0x80 : s * 0x7F;
        }
        return output;
    }
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

    // End audio exp.
});
