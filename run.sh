#!/usr/bin/env bash
# Starts/stops all 5 ChitNya services on fixed ports that stay out of the way
# of other projects on this Mac (which tend to grab 8080/8081/5173).
#
#   ./run.sh            start everything
#   ./run.sh stop       stop everything
#   ./run.sh status     show what's up
#   ./run.sh restart    stop then start

set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"
ROOT="$PWD"
RUN_DIR="$ROOT/.run"
mkdir -p "$RUN_DIR"

API_PORT=8090
ADMIN_PORT=8092
BRIDGE_PORT=9001
SITE_PORT=5273
ADMIN_UI_PORT=5274

port_owner() { lsof -nP -iTCP:"$1" -sTCP:LISTEN -t 2>/dev/null; }

lan_ip() { ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null; }

start_one() {
  local name="$1" port="$2"; shift 2
  local pidfile="$RUN_DIR/$name.pid" logfile="$RUN_DIR/$name.log"
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
    echo "  $name: already running (pid $(cat "$pidfile"))"
    return
  fi
  local existing; existing="$(port_owner "$port")"
  if [ -n "$existing" ]; then
    echo "  $name: SKIPPED — port $port is used by another process (pid $existing, not ours)"
    return
  fi
  ( "$@" >"$logfile" 2>&1 & echo $! >"$pidfile" )
  sleep 1
  echo "  $name: started (pid $(cat "$pidfile"), log .run/$name.log)"
}

do_start() {
  if ! pgrep -x mongod >/dev/null 2>&1 && ! pgrep -f "mongod" >/dev/null 2>&1; then
    echo "warning: mongod doesn't look like it's running — start it first (e.g. 'brew services start mongodb-community')"
  fi

  echo "building backend..."
  ( cd backend && go build -o "$RUN_DIR/api-bin" ./cmd/api && go build -o "$RUN_DIR/admin-bin" ./cmd/admin ) || { echo "backend build failed"; exit 1; }

  echo "starting services..."
  ( cd backend && PORT=$API_PORT start_one api "$API_PORT" "$RUN_DIR/api-bin" )
  ( cd backend && ADMIN_PORT=$ADMIN_PORT start_one admin "$ADMIN_PORT" "$RUN_DIR/admin-bin" )
  if [ -f bridge/config.ini ] && [ -x bridge/.venv/bin/python ]; then
    ( cd bridge && start_one bridge "$BRIDGE_PORT" .venv/bin/python server.py )
  else
    echo "  bridge: SKIPPED — bridge/config.ini or bridge/.venv missing (only needed for adding Telegram videos)"
  fi
  ( cd frontend && API_PROXY="http://localhost:$API_PORT" start_one frontend "$SITE_PORT" npm run dev -- --port "$SITE_PORT" --strictPort --host )
  ( cd admin && ADMIN_API="http://localhost:$ADMIN_PORT" start_one adminui "$ADMIN_UI_PORT" npm run dev -- --port "$ADMIN_UI_PORT" --strictPort --host )

  echo
  echo "waiting for services to come up..."
  sleep 3
  do_status
}

do_stop() {
  echo "stopping services..."
  for name in api admin bridge frontend adminui; do
    local pidfile="$RUN_DIR/$name.pid"
    if [ -f "$pidfile" ]; then
      local pid; pid="$(cat "$pidfile")"
      if kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null
        echo "  $name: stopped (pid $pid)"
      else
        echo "  $name: not running"
      fi
      rm -f "$pidfile"
    else
      echo "  $name: not running"
    fi
  done
}

check_health() {
  local label="$1" url="$2"
  if curl -s -m 2 -o /dev/null "$url" 2>/dev/null; then
    printf "  %-12s UP    %s\n" "$label" "$url"
  else
    printf "  %-12s down  %s\n" "$label" "$url"
  fi
}

do_status() {
  echo "--- status ---"
  check_health "public API"  "http://localhost:$API_PORT/health"
  check_health "admin API"   "http://localhost:$ADMIN_PORT/health"
  check_health "bridge"      "http://localhost:$BRIDGE_PORT/health"
  check_health "user site"   "http://localhost:$SITE_PORT/"
  check_health "admin panel" "http://localhost:$ADMIN_UI_PORT/"
  local ip; ip="$(lan_ip)"
  echo
  echo "  user site (this Mac):  http://localhost:$SITE_PORT"
  [ -n "$ip" ] && echo "  user site (phone, same Wi-Fi): http://$ip:$SITE_PORT"
  echo "  admin panel:            http://localhost:$ADMIN_UI_PORT"
}

case "${1:-start}" in
  start)   do_start ;;
  stop)    do_stop ;;
  restart) do_stop; sleep 1; do_start ;;
  status)  do_status ;;
  *) echo "usage: $0 [start|stop|restart|status]"; exit 1 ;;
esac
