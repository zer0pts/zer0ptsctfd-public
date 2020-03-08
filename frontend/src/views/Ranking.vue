<template>
  <table class="table ranking">
    <thead>
      <tr>
        <th>Rank</th>
        <th colspan="2">Team</th>
        <th>Score</th>
        <th v-for="c in orderedChallenges" :key="c.id" class="challenge-name">
          <span>
            {{ c.name }}
          </span>
        </th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="t in teamOrderByRank" :key="t.id">
        <th>{{ t.rank }}</th>
        <th style="width: 1em;">
          <CountryFlag
            :country="t.country_code"
            v-if="t.country_code"
          ></CountryFlag>
        </th>
        <th>
          <router-link :to="{ name: 'Team', params: { id: t.id } }">
            {{ t.teamname }}
          </router-link>
        </th>
        <th>{{ t.score }}</th>
        <td v-for="c in orderedChallenges" :key="c.id">
          <b-icon
            icon="flag"
            type="is-black"
            v-if="teamSolvedChallenge(t, c)"
          />
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script>
import API from "../api";
import { handleError } from "../util";
import lodash from "lodash";
import CountryFlag from "vue-country-flag";

export default {
  components: {
    CountryFlag: CountryFlag
  },
  data() {
    return {
      teams: [],
      challenges: []
    };
  },
  methods: {
    teamSolvedChallenge(t, c) {
      return c.solveteams.includes(t.id);
    }
  },
  mounted() {
    API.get("/teams")
      .then(r => {
        this.teams = r.data.teams;
      })
      .catch(e => handleError(this, e));

    API.get("/challenges")
      .then(r => {
        this.challenges = lodash.keyBy(r.data.challenges, "id");
      })
      .catch(e => handleError(this, e));

    this.$eventHub.$on("challengeUpdate", c => {
      this.$set(this.challenges, c.id, c);
      this.$forceUpdate();
      this.teams = Object.assign({}, this.teams);
    });
    this.$eventHub.$on("challengeClose", cid => {
      delete this.challenges[cid];
      this.$forceUpdate();
    });
  },
  computed: {
    orderedChallenges() {
      return lodash(this.challenges)
        .orderBy(["category", "name"])
        .value();
    },
    teamOrderByRank() {
      let teams = lodash(this.teams)
        .map(t => {
          return {
            score: lodash(this.challenges)
              .filter(c => c.solveteams.includes(t.id))
              .map("score")
              .sum(),
            last_submission: lodash.max(
              t.submissions.map(s => {
                if (
                  this.challenges[s.challenge_id] &&
                  this.challenges[s.challenge_id].is_questionary
                ) {
                  return 0;
                }
                return s.submitted_at;
              })
            ),
            ...t
          };
        })
        .orderBy(["score", "last_submission"], ["desc", "asc"])
        .map((t, i) => {
          return {
            rank: i + 1,
            ...t
          };
        })
        .value();
      for (let i = 1; i < teams.length; i++) {
        if (
          teams[i].score === teams[i - 1].score &&
          teams[i].last_submission === teams[i - 1].last_submission
        ) {
          teams[i].rank = teams[i - 1].rank;
        }
      }
      return teams;
    }
  }
};
</script>

<style lang="scss" scoped>
.ranking {
  width: 100%;
  border-collapse: separate;
  padding-top: 10em;
}
.challenge-name {
  span {
    display: inline-block;
    transform-origin: left bottom;
    transform: rotate(-45deg);
    width: 1em;
  }
  white-space: pre;
}
</style>
