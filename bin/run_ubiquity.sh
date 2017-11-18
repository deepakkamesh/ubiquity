#!/bin/sh
LOC=$(dirname "$0")
killall main

# Delete old logs.
find $LOC/../logs -mindepth 1 -type f -mtime +10 -delete

$LOC/main \
-log_dir=$LOC/../logs/ \
-resources=$LOC/../resources \
-stderrthreshold=info \
-logtostderr=true \
-enable_pi_gpio=false \
-http_port="10.138.0.2:80" \
-enable_video=false \
-enable_audio=false \
-ssl_cert="server.crt" \
-ssl_priv_key="server.key" \
-v=1
