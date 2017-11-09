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

    ws = new WebSocket("ws://" + window.location.host + "/datastream");

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
                Param: "",
            }
        }));
    });

});
