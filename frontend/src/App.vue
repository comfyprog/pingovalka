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
    }
  },

  methods: {
    
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
    <div class="columns mt-2">
      <div class="column notification is-success is-light m-1" v-for="host in hosts" :key="host.id">
        <h1 class="title">{{ host.name }}</h1>
        <h2 class="subtitle">{{ host.addr }}</h2>
      </div>
    </div>
  </template>
</template>

