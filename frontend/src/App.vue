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
      soundUp: null,
      soundUpPlay: false,
      soundDown: null,
      soundDownPlay: false,
    }
  },

  methods: {
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
      console.log(new Date(), `host ${host.name} [${host.addr}] changed status to '${host.status}'`);
      for (let i = 0; i < this.hosts.length; i++) {
        if (this.hosts[i].id == host.id) {
          this.hosts[i].status = host.status;
          this.playStatusSound(host.status);
          return;
        }
      }
    }
    
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
    this.soundUp = new Audio("/sounds/up.wav");
    this.soundDown = new Audio("/sounds/down.wav");
    let socket = new WebSocket(this.wsUrl);
    this.socket = socket;

    socket.onopen = function() {
    }

    socket.onclose = event => {
      if (event.wasClean) {
        console.log("clean socket close: ", event);
      } else {
        console.log("dirty socket close: ", event);
      }
      this.currentState = this.State.Error;
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
}
</script>

<template>
  <template v-if="currentState == State.Error">
  Connection error, try reloading the page.
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
      <div class="column notification is-light m-1" :class="getHostClass(host)" v-for="host in row" :key="host.id">
        <h1 class="title" :class="titleClass">
          <template v-if="host.status == 'online'">
          ↑
          </template>
          <template v-if="host.status == 'offline'">
          ↓
          </template>
          {{ host.name }}
        </h1>
        <h2 class="subtitle">{{ host.addr }}</h2>
      </div>

      <div class="column m-1" v-for="n in hostsPerRow - row.length" :key="n">
      </div>
    </div>

  </template>
</template>

