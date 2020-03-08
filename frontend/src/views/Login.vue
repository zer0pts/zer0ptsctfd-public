<template>
  <form class="column is-half is-offset-one-quarter" @submit.prevent="login">
    <b-field label="username">
      <b-input v-model="username"></b-input>
    </b-field>
    <b-field label="password">
      <b-input type="password" password-reveal v-model="password"></b-input>
    </b-field>
    <div class="has-text-right">
      <b-button tag="input" native-type="submit" value="Login"></b-button>
      <br />
      <router-link to="/reset-request">password reset</router-link>
    </div>
  </form>
</template>

<script>
import API from "../api";
import { handleError } from "../util";
export default {
  data() {
    return {
      username: "",
      password: ""
    };
  },
  methods: {
    login() {
      API.post("/login", {
        username: this.username,
        password: this.password
      })
        .then(r => {
          if (r.data.message) {
            this.$buefy.snackbar.open({
              message: r.data.message,
              queue: false
            });
          }
          this.$eventHub.$emit("checkLogin");
          this.$router.push("/");
        })
        .catch(e => {
          handleError(this, e);
        });
    }
  }
};
</script>
