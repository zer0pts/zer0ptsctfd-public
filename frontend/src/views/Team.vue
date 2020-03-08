<template>
  <section v-if="loaded">
    <h2 class="title is-2">
      <CountryFlag
        :country="team.country_code"
        v-if="team.country_code"
      ></CountryFlag>
      {{ team.teamname }}
      <span class="is-size-4"> [{{ teamScore }}]</span>
    </h2>

    <section v-if="team.token" class="column is-offset-one">
      <b-field label="team token">
        <b-input :value="team.token" readonly />
      </b-field>
    </section>

    <section v-if="team.token" class="column is-offset-one">
      <b-field label="teamname">
        <b-input v-model="teamname"></b-input>
      </b-field>
      <div class="is-clearfix">
        <div class="is-pulled-right buttons">
          <b-button @click="setTeamName">update</b-button>
        </div>
      </div>
    </section>

    <section v-if="team.token" class="column is-offset-one">
      <label class="label">country code</label>
      <div class="control has-icons-left">
        <input class="input" type="text" v-model="country" />
        <span class="icon is-small is-left">
          <CountryFlag :country="country"></CountryFlag>
        </span>
      </div>
      <div class="is-clearfix">
        <div class="is-pulled-right buttons">
          <b-button @click="setCountry">update</b-button>
        </div>
      </div>
    </section>

    <section class="column is-offset-one">
      <div class="box member" v-for="u in team.users" :key="u.id">
        {{ u.username }} [{{ userScore(u) }}]
      </div>
    </section>

    <div class="timeline">
      <div class="timeline-item" v-for="(s, i) in submissions" :key="i">
        <div class="timeline-marker is-primary"></div>
        <div class="timeline-content" v-if="challenges[s.challenge_id]">
          <p class="heading">
            [{{ submissionTime(s) }}] {{ challenges[s.challenge_id].name }}
          </p>
          <p>
            {{ submitUser(s) }} earned
            {{ challenges[s.challenge_id].score }} points
          </p>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
import CountryFlag from "vue-country-flag";
import API from "../api";
import { handleError, showMessage } from "../util";
import lodash from "lodash";
import dayjs from "dayjs";

export default {
  components: {
    CountryFlag: CountryFlag
  },
  data() {
    return {
      team: null,
      teamname: "",
      country: null,
      challenges: []
    };
  },
  methods: {
    userScore(u) {
      return lodash.sum(
        this.team.submissions
          .filter(s => s.user_id === u.id)
          .map(s =>
            this.challenges[s.challenge_id]
              ? this.challenges[s.challenge_id].score
              : 0
          )
      );
    },
    submitUser(s) {
      return this.team.users.filter(u => u.id === s.user_id)[0].username;
    },
    submissionTime(s) {
      return dayjs(s.submitted_at * 1000).format("YYYY-MM-DD HH:mm:ss Z");
    },
    getTeam() {
      const id = this.$route.params.id;
      API.get("/team/" + id)
        .then(r => {
          this.team = r.data.team;
          this.teamname = r.data.team.teamname;
          this.country = r.data.team.country_code;
        })
        .catch(e => handleError(this, e));
    },
    setTeamName() {
      API.post("/set-teamname", {
        teamname: this.teamname
      })
        .then(r => {
          showMessage(this, r.data.message);
          window.location.reload();
        })
        .catch(e => handleError(this, e));
    },
    setCountry() {
      API.post("/set-country", {
        country: this.country
      })
        .then(r => {
          showMessage(this, r.data.message);
          window.location.reload();
        })
        .catch(e => handleError(this, e));
    }
  },
  mounted() {
    this.getTeam();
    API.get("/challenges")
      .then(r => {
        this.challenges = r.data.challenges.reduce((map, c) => {
          map[c.id] = c;
          return map;
        }, {});
      })
      .catch(e => handleError(this, e));
  },
  computed: {
    teamScore() {
      return lodash.sum(
        this.team.submissions.map(s =>
          this.challenges[s.challenge_id]
            ? this.challenges[s.challenge_id].score
            : 0
        )
      );
    },
    submissions() {
      return lodash.sortBy(this.team.submissions, "submitted_at");
    },
    loaded() {
      return this.challenges && this.team;
    }
  }
};
</script>

<style lang="scss" scoped>
.flag {
  vertical-align: bottom;
}
.member {
  display: inline-flex;
  padding: 0.75rem;
  margin: 0.25rem 0.75rem;
  border-radius: 0;
}
</style>
