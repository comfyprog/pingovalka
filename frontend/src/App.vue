<script>

const State =  {
  Loading: `Loading`,
  Ready: `Ready`,
  Error: `Error`,
};

export default {
  data() {
    return {
      State,
      currentState: State.Loading,
      wsUrl: wsUrl,
      socket: null,
      hosts: [],
      hostsPerRow: 5,
      fontSize: 5,
      soundUp: null,
      soundUpPlay: false,
      soundDown: null,
      soundDownPlay: false,
      reconnectInterval: 0,
      maxReconnectInterval: 10000,
      reconnectTimerId: null,
      hostInfo: null,
    }
  },

  methods: {
    getWebsocketAddress() {
      if (location.protocol.startsWith("https")) {
        return `wss://${location.host}${wsUrl}`
      }
        return `ws://${location.host}${wsUrl}`
    },
  
    playStatusSound(status) {
      if (status == "online" && this.soundUpPlay) {
        this.soundUp.play();
      }
      if (status == "offline" && this.soundDownPlay) {
        this.soundDown.play();
      }
    },
    increaseHostsPerRow() {
      this.hostsPerRow = Math.min(15, this.hostsPerRow+1);
    },
    decreaseHostsPerRow() {
      this.hostsPerRow = Math.max(1, this.hostsPerRow-1);
    },


    decreaseFontSize() {
      this.fontSize = Math.min(6, this.fontSize+1);
    },
    increaseFontSize() {
      this.fontSize = Math.max(1, this.fontSize-1);
    },
    
    getHostClass(host) {
      if (host.status == "offline") {
        return "is-danger";
      } else if (host.status == "online") {
        return "is-success";
      } else {
        return "is-primary";
      }
    },

    changeStatus(host) {      
      console.log(
        this.convertUnixSecondsToDateString(host.statusChangeTime),
        `host ${host.name} [${host.addr}] changed status to '${host.status}'`);
      for (let i = 0; i < this.hosts.length; i++) {
        if (this.hosts[i].id == host.id) {
          this.hosts[i].status = host.status;
          this.hosts[i].statusChangeTime = host.statusChangeTime;
          this.playStatusSound(host.status);
          return;
        }
      }
    },

    convertUnixSecondsToDateString(n) {
      return (new Date(n * 1000)).toLocaleString();
    },

    resetReconnectTimer() {
      if (!this.reconnectTimerId) {
        return;
      }
      clearTimeout(this.reconnectTimerId);
      this.reconnectTimerId = null;
    },

    reconnect() {
      if (this.reconnectTimerId) {
        return;
      }
      this.reconnectInterval = Math.min(this.maxReconnectInterval, this.reconnectInterval + 1000);
      console.log(`attempting to reconnect in ${this.reconnectInterval}ms`);
      window.reconnectTimerId = setTimeout(() => this.interactWithServer(), this.reconnectInterval);
    },

    interactWithServer() {
      let wsAddress = this.getWebsocketAddress();
      let socket = new WebSocket(wsAddress);
      this.socket = socket;

      socket.onerror = event => {
        console.log(`websocket error: ${event}`);
        this.resetReconnectTimer();
        this.reconnect();
      }

      socket.onopen = () => {
        console.log("established new websocket connection")
        this.resetReconnectTimer();
        this.reconnectInterval = 0;
      }

      socket.onclose = event => {
        if (event.wasClean) {
          console.log("clean socket close: ", event);
        } else {
          console.log("dirty socket close: ", event);
        }
        this.currentState = this.State.Error;

        this.reconnect();
      }

      socket.onerror = event => {
        console.log("socket error", event);
        this.currentState = this.State.Error;
      }

      socket.onmessage = event => {
        let data = JSON.parse(event.data);
        if (data.type == "list") {
          this.hosts = data.data;
          this.currentState = this.State.Ready;
        }

        if (data.type == "status") {
          this.changeStatus(data.data);
        }
      }

    },
    
  },

  computed: {
    hostsRows() {
      let rows = [];
      let row = [];

      for (let host of this.hosts) {
        row.push(host);
        if (row.length == this.hostsPerRow) {
          rows.push(row);
          row = [];
        }
      }

      if (row.length > 0) {
        rows.push(row);
      }

      return rows;
    },

    titleClass() {
      return `is-${this.fontSize}`;
    },
  },

  mounted() {
    this.soundUp = new Audio("/sounds/up.wav");
    this.soundDown = new Audio("/sounds/down.wav");

    this.interactWithServer();
  },
}
</script>

<template>
  <template v-if="currentState == State.Error">
  Connection error, will try to reconnect in {{ Math.floor(reconnectInterval / 1000) }}s.
  </template>

  <template v-if="currentState == State.Loading">
  Loading...
  </template>

  <template v-if="currentState == State.Ready">
    <nav class="navbar m-2" role="navigation" aria-label="navigation">
      <div class="navbar-menu is-active is-flex is-vcentered" id="navbar">
        <div class="navbar-start">
          <span>
          <a href="#" @click.prevent="increaseHostsPerRow()">[+]</a>
          <a href="#" @click.prevent="decreaseHostsPerRow()">[-]</a>
          hosts per row
          </span>



          <span>
          <a href="#" @click.prevent="increaseFontSize()">[+]</a>
          <a href="#" @click.prevent="decreaseFontSize()">[-]</a>
          font size
          </span>

          <label class="checkbox m-1">
            <input type="checkbox" v-model="soundDownPlay">
            sound on host down
          </label>


          <label class="checkbox m-1">
            <input type="checkbox" v-model="soundUpPlay">
            sound on host up
          </label>

        </div>
      </div>
    </nav>


  
    <div class="columns " v-for="(row, rowIndex) in hostsRows" :key="rowIndex">
      <div class="column notification is-light m-1" :class="getHostClass(host)" v-for="host in row" :key="host.id"
        @click="hostInfo = host">
        <h1 class="title" :class="titleClass">
          <template v-if="host.status == 'online'">
          ↑
          </template>
          <template v-if="host.status == 'offline'">
          ↓
          </template>
          {{ host.name }}
        </h1>
        <h6 class="subtitle is-7">
        {{ host.addr }}
        </h6>
      </div>

      <div class="column m-1" v-for="n in hostsPerRow - row.length" :key="n">
      </div>
    </div>



    <div class="modal" :class="{ 'is-active': hostInfo }" v-if="hostInfo">
      <div class="modal-background" @click="hostInfo = null"></div>
      <div class="modal-content">
        <div class="card">
          <div class="card-content">
            <p class="title">{{ hostInfo.name }}</p>
            <p class="subtitle">{{ hostInfo.addr }}</p>

            <div class="content">
              <div class="notification is-light" :class="{'is-danger': hostInfo.status == 'offline', 'is-success': hostInfo.status == 'online'}">
                <small>
                 <strong>{{ hostInfo.status }} since: {{ convertUnixSecondsToDateString(hostInfo.statusChangeTime) }}</strong>
                </small>
              </div>

              <template v-if="hostInfo.info">
                <table class="table is-bordered is-hoverable is-narrow is-fullwidth">
                <template v-for="info in hostInfo.info">
                  <tr><th class="title is-6 has-text-centered">{{ info.title }}</th></tr>
                  <tr>
                    <td class="subtitle is-6" v-if="info.isHtml" v-html="info.text"></td>
                    <td class="subtitle is-6" v-else>{{ info.text }}</td>
                  </tr>
                </template>
                </table>
              
              </template>
              
            </div>
          </div>

        </div>
      </div>

      <button class="modal-close is-large" aria-label="close" @click="hostInfo = null"></button>
    </div>
    
  </template>
</template>

