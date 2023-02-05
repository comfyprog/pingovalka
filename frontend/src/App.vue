<script>

const State =  {
  Loading: `Loading`,
  Ready: `Ready`,
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
    }
  },

  methods: {
    increaseHostsPerRow() {
      this.hostsPerRow = Math.min(15, this.hostsPerRow+1);
    },
    decreaseHostsPerRow() {
      this.hostsPerRow = Math.max(1, this.hostsPerRow-1);
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
      if (this.hostsPerRow <= 5) {
        return "is-2";
      } else if (this.hostsPerRow <= 10) {
        return "is-3";
      } else {
        return "is-4";
      }
    },
  },

  mounted() {
    this.wsUrl = `ws://localhost:9999/ws`;
    let socket = new WebSocket(this.wsUrl);
    this.socket = socket;

    socket.onopen = function(e) {
      console.log("socket opened", e);
    }

    socket.onclose = function(event) {
      if (event.wasClean) {
        console.log("socket closed cleanly", event);
      } else {
        console.log("socket closed dirty");
      }
    }

    socket.onerror = function(event) {
      console.log("socket error", event);
    }

    socket.onmessage = (event) => {
      console.log("socket message", event)
      let data = JSON.parse(event.data);
      if (data.type == "list") {
        this.hosts = data.data;
        console.log(this.hosts);
        this.currentState = this.State.Ready;
      }
    }
    
    
  },
}
</script>

<template>
  <template v-if="currentState == State.Loading">
  Loading...
  </template>

  <template v-if="currentState == State.Ready">
    <nav class="navbar m-2" role="navigation" aria-label="navigation">
      <div class="navbar-menu is-active" id="navbar">
        <div class="navbar-start">
          <button class="button" @click="increaseHostsPerRow()">-</button>
          <button class="button" @click="decreaseHostsPerRow()">+</button>

        </div>
      </div>
    </nav>


  
    <div class="columns " v-for="(row, rowIndex) in hostsRows" :key="rowIndex">
      <div class="column notification is-light m-1" :class="getHostClass(host)" v-for="host in row" :key="host.id">
        <h1 class="title" :class="titleClass">
          {{ host.name }}
          <template v-if="host.status == 'online'">
          ↑
          </template>
          <template v-if="host.status == 'offline'">
          ↓
          </template>
        </h1>
        <h2 class="subtitle">{{ host.addr }}</h2>
      </div>

      <div class="column m-1" v-for="n in hostsPerRow - row.length" :key="n">
      </div>
    </div>

  </template>
</template>

