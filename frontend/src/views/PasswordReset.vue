<template>
  <form class="column is-half is-offset-one-quarter" @submit.prevent="request">
    <b-field label="token">
      <b-input v-model="token"></b-input>
    </b-field>
    <b-field label="new password">
      <b-input type="password" password-reveal v-model="password"></b-input>
    </b-field>
    <div class="has-text-right">
      <b-button
        tag="input"
        native-type="submit"
        value="Reset Password"
      ></b-button>
    </div>
  </form>
</template>

<script>
import API from "../api";
import { handleError } from "../util";
export default {
  data() {
    return {
      token: "",
      password: ""
    };
  },
  methods: {
    request() {
      API.post("/reset", {
        token: this.token,
        password: this.password
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
        .catch(e => {
          handleError(this, e);
        });
    }
  }
};
</script>
