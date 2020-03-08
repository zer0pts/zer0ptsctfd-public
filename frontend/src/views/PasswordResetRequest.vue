<template>
  <form class="column is-half is-offset-one-quarter" @submit.prevent="request">
    <b-field label="email">
      <b-input v-model="email"></b-input>
    </b-field>
    <div class="has-text-right">
      <b-button
        tag="input"
        native-type="submit"
        value="Request Password Reset"
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
      email: ""
    };
  },
  methods: {
    request() {
      API.post("/reset-request", {
        email: this.email
      })
        .then(r => {
          if (r.data.message) {
            this.$buefy.snackbar.open({
              message: r.data.message,
              queue: false
            });
          }
          this.$router.push("/reset");
        })
        .catch(e => {
          handleError(this, e);
        });
    }
  }
};
</script>
