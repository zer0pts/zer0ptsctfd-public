<template>
  <form class="column is-half is-offset-one-quarter">
    <b-field label="username">
      <b-input v-model="username"></b-input>
    </b-field>
    <b-field label="email">
      <b-input type="email" v-model="email"></b-input>
    </b-field>
    <b-field label="password">
      <b-input type="password" password-reveal v-model="password"></b-input>
    </b-field>
    <div class="is-clearfix">
      <div class="is-pulled-right buttons">
        <b-button @click="mode = 'create'">Create new Team</b-button>
        <b-button @click="mode = 'join'">Join existing Team</b-button>
      </div>
    </div>

    <template v-if="mode === 'create'">
      <b-field label="teamname">
        <b-input v-model="teamname"></b-input>
      </b-field>

      <div class="field">
        <label class="label">country code</label>
        <div class="control has-icons-left">
          <input class="input" type="text" v-model="country" />
          <span class="icon is-small is-left">
            <CountryFlag :country="country"></CountryFlag>
          </span>
        </div>
      </div>

      <div class="is-pulled-right buttons">
        <b-button @click="create">Create</b-button>
      </div>
    </template>

    <template v-if="mode === 'join'">
      <b-field label="teamtoken">
        <b-input v-model="teamtoken"></b-input>
      </b-field>
      <div class="is-pulled-right buttons">
        <b-button @click="jointeam">Join</b-button>
      </div>
    </template>
  </form>
</template>

<script>
import CountryFlag from "vue-country-flag";
import API from "../api";
import { handleError } from "../util";

export default {
  components: {
    CountryFlag: CountryFlag
  },
  data() {
    return {
      mode: "register",
      username: "",
      email: "",
      password: "",

      teamname: "",
      country: "JPN",

      teamtoken: ""
    };
  },
  methods: {
    jointeam() {
      API.post("/join-team", {
        username: this.username,
        email: this.email,
        password: this.password,
        token: this.teamtoken
      })
        .then(r => {
          if (r.data.message) {
            this.$buefy.snackbar.open({
              message: r.data.message,
              queue: false
            });
          }
          this.$router.push("/login");
        })
        .catch(e => handleError(this, e));
    },
    create() {
      API.post("/create-team", {
        username: this.username,
        email: this.email,
        password: this.password,
        teamname: this.teamname,
        country: this.country
      })
        .then(r => {
          if (r.data.message) {
            this.$buefy.snackbar.open({
              message: r.data.message,
              queue: false
            });
          }
          this.$router.push("/login");
        })
        .catch(e => handleError(this, e));
    }
  }
};
</script>

<style lang="stylus" scoped></style>
