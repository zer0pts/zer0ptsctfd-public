<template>
  <div class="container-fluid">
    <b-navbar type="is-primary">
      <template slot="brand">
        <b-navbar-item
          tag="router-link"
          to="/"
          class="is-size-3 has-text-weight-bold"
          >{{ ctfName }}</b-navbar-item
        >
      </template>
      <template slot="start">
        <template v-if="login">
          <b-navbar-item tag="router-link" to="/challenges"
            >Challenges</b-navbar-item
          >
          <b-navbar-item tag="router-link" to="/ranking">Ranking</b-navbar-item>
        </template>
      </template>

      <template slot="end">
        <b-navbar-item tag="div">
          <b-tooltip label="Reconnect WebSocket" position="is-bottom">
            <b-icon
              icon="flash"
              :type="wsStatus ? 'is-black' : 'is-danger'"
              tooltip
              @click.native="wsConnect"
            >
            </b-icon>
          </b-tooltip>
        </b-navbar-item>
        <b-navbar-item tag="div" class="buttons">
          <template v-if="login">
            <b-button>{{ username }}</b-button>
            <b-button
              tag="router-link"
              v-if="team"
              :to="{ name: 'Team', params: { id: team.id } }"
              >{{ team.teamname }}</b-button
            >
            <b-button @click="logout">logout</b-button>
          </template>
          <template v-else>
            <b-button tag="router-link" to="/login">login</b-button>
            <b-button tag="router-link" to="/register">register</b-button>
          </template>
        </b-navbar-item>
      </template>
    </b-navbar>
    <div class="container overflower">
      <router-view></router-view>
    </div>
  </div>
</template>

<script>
import API from "./api";
import { SERVER_ADDRESS } from "./env";
import { handleError } from "./util";

export default {
  data() {
    return {
      ctfName: "",
      wsStatus: false,
      ws: null,
      username: null,
      team: null
    };
  },
  methods: {
    logout() {
      API.post("/logout")
        .then(() => {
          this.username = null;
          this.team = null;
          this.$router.push("/");
        })
        .catch(e => handleError(this, e));
    },
    wsConnect() {
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }

      this.ws = new WebSocket(SERVER_ADDRESS.replace(/^http/, "ws") + "/ws");
      this.ws.addEventListener("open", () => {
        this.wsStatus = true;
      });
      this.ws.addEventListener("close", () => {
        this.wsStatus = false;
      });
      this.ws.addEventListener("error", () => {
        this.wsStatus = false;
      });
      this.ws.addEventListener("message", e => {
        const data = JSON.parse(e.data);
        this.$eventHub.$emit(data.type, data.value);
      });
    },
    checkLogin() {
      API.get("/user").then(r => {
        this.username = r.data.username;

        return API.get("/team").then(r => {
          this.team = r.data.team;
        });
      });
    }
  },
  mounted() {
    this.wsConnect();
    API.get("/ctf")
      .then(r => {
        if (r.data.config) {
          ({ ctf_name: this.ctfName } = r.data.config);
          document.title = this.ctfName;
        }
      })
      .catch(e => {
        this.$buefy.snackbar.open({
          message: e.response.data.message
            ? e.response.data.message
            : "internal server error",
          type: "is-warning"
        });
      });

    this.$eventHub.$on("message", msg => {
      this.$buefy.snackbar.open({
        message: msg,
        type: "is-primary",
        queue: false
      });
    });
    this.checkLogin();
    this.$eventHub.$on("checkLogin", () => {
      this.checkLogin();
    });
  },
  computed: {
    login() {
      return this.username;
    }
  }
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

#nav {
  padding: 30px;
}

#nav a {
  font-weight: bold;
  color: #2c3e50;
}

#nav a.router-link-exact-active {
  color: #42b983;
}
.overflower {
  overflow: auto;
}
</style>
