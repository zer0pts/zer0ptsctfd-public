<template>
  <form class="column is-half is-offset-one-quarter" @submit.prevent>
    <b-field label="CTF Name">
      <b-input v-model="ctfName"></b-input>
    </b-field>

    <b-field label="CTF Start At">
      <b-datetimepicker
        v-model="startAt"
        placeholder="Click to select..."
        :datetime-formatter="datetimeFormatter"
        :datetime-parset="datetimeParser"
        editable
      >
        <template slot="left">
          <button class="button is-primary" @click="startAt = new Date()">
            <b-icon icon="clock"></b-icon>
            <span>Now</span>
          </button>
        </template>
      </b-datetimepicker>
    </b-field>

    <b-field label="CTF End At">
      <b-datetimepicker
        v-model="endAt"
        placeholder="Click to select..."
        :datetime-formatter="datetimeFormatter"
        :datetime-parset="datetimeParser"
        editable
      >
        <template slot="left">
          <b-button type="is-primary" @click="endAt = new Date()">
            Now
          </b-button>
        </template>
      </b-datetimepicker>
    </b-field>

    <b-field label="Submission Lock">
      <p>
        Lock Submission:
        <b-input v-model="lockDuration" type="number" /> seconds if team
        submitted <b-input v-model="lockCount" type="number" /> wrong flags
        consecutive in <b-input v-model="lockSecond" type="number" /> seconds.
      </p>
    </b-field>

    <b-field label="Estimate number of easy challenge solves">
      <b-input v-model="easySolves" type="number" />
    </b-field>

    <b-field label="Estimate number of medium challenge solves">
      <b-input v-model="mediumSolves" type="number" />
    </b-field>

    <b-field label="Minimum Score of challenge">
      <b-input v-model="minScore" type="number" />
    </b-field>

    <div class="is-clearfix">
      <div class="is-pulled-right buttons">
        <b-button type="is-warning" @click="getValues">reset</b-button>
        <b-button type="is-primary" @click="setValues">update</b-button>
      </div>
    </div>
  </form>
</template>

<script>
import API from "../../api";
import { handleError } from "../../util";
import dayjs from "dayjs";

export default {
  data() {
    return {
      ctfName: "",
      startAt: null,
      endAt: null,
      lockSecond: 0,
      lockDuration: 0,
      lockCount: 0,
      easySolves: 0,
      mediumSolves: 0,
      minScore: 0
    };
  },
  methods: {
    datetimeFormatter(d) {
      return dayjs(d).format("YYYY-MM-DD HH:mm:ss Z");
    },
    datetimeParser(s) {
      return new Date(dayjs(s, "YYYY-MM-DD HH:mm:ss Z").valueOf());
    },
    getValues() {
      API.get("/ctf")
        .then(r => {
          if (r.data) {
            let start_at, end_at;
            ({
              ctf_name: this.ctfName,
              start_at,
              end_at,
              lock_second: this.lockSecond,
              lock_count: this.lockCount,
              lock_duration: this.lockDuration,
              easy_solves: this.easySolves,
              medium_solves: this.mediumSolves,
              min_score: this.minScore
            } = r.data.config);
            this.startAt = new Date(start_at * 1000);
            this.endAt = new Date(end_at * 1000);
          }
        })
        .catch(e => handleError(this, e));
    },
    setValues() {
      API.post("/set-ctf", {
        ctf_name: this.ctfName,
        start_at: Math.floor(this.startAt.valueOf() / 1000),
        end_at: Math.floor(this.endAt.valueOf() / 1000),
        lock_second: +this.lockSecond,
        lock_count: +this.lockCount,
        lock_duration: +this.lockDuration,
        easy_solves: +this.easySolves,
        medium_solves: +this.mediumSolves,
        min_score: +this.minScore
      })
        .then(r => {
          this.$buefy.snackbar.open({
            message: r.data.message,
            queue: false
          });
        })
        .catch(e => handleError(this, e));
    }
  },
  mounted() {
    this.getValues();
  }
};
</script>
