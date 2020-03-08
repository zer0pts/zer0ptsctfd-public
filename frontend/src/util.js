export const handleError = (that, e) => {
  let message = "";
  if (!e.response) {
    message = "Network Error";
  } else {
    message = e.response.data.message || "internal server error";
  }
  that.$buefy.snackbar.open({
    message: message,
    type: "is-warning",
    queue: false
  });
};

export const showMessage = (that, m) => {
  that.$buefy.snackbar.open({
    message: m,
    queue: false
  });
};
