<?php
$pdo = new PDO("mysql:host=db;dbname=login;", "user", "password");
$pdo->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);

$login_user = null;

if (isset($_POST["login"]) && isset($_POST["username"]) && isset($_POST["password"])) {
    $rows = $pdo->query("SELECT username FROM users WHERE username='${_POST['username']}' AND password='${_POST['password']}' LIMIT 1");
    foreach ($rows as $row) {
        $login_user = $row['username'];
        break;
    }
}
else if (isset($_POST["register"]) && isset($_POST["username"]) && isset($_POST["password"])){
    $stmt = $pdo->prepare("INSERT INTO users(username, password) values(:username, :password)");
    $stmt->execute([
        ":username" => $_POST["username"],
        ":password" => $_POST["password"],
    ]);
}
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>JUST LOGIN</title>

<style>
body {
    background-color: gray;
}
#container{
    width: 800px;
    position: absolute;
    top: 50%;
    left: 50%;
    margin-left: -25%;
    border: 10px solid black;
}
form {
    padding: 20px 5px;
}
input[type=text],input[type=password] {
    display: block;
    width: 100%;
    background: white;
    border: 2px solid black;
    color: black;
    font-size: 26px;
    padding: 2px 5px;
    text-align: center;
}
input[type=submit] {
    padding: 2px 4px;
    border: 2px solid black;
    background: black;
    color: white;
    font-size: 26px;
    text-align: center;
}
.twocolumn {
    display: flex;
}
.message {
    text-align: center;
    color: white;
    font-size: 32px;
    font-weight: bold;
    font-family: sans-serif;
}

</style>
</head>
<body>
    <div id="container">
        <?php if ($login_user === "admin") { ?>
            <div class="message"><flag></div>
        <?php } else if ($login_user) { ?>
            <div class="message"><?= $login_user ?></div>
        <?php } else { ?>
            <div class="twocolumn">
                <form action="index.php" method="POST">
                    <div><input type="text" name="username" placeholder="USERNAME"></div>
                    <div><input type="password" name="password" placeholder="PASSWORD"></div>
                    <div><input type="submit" name="register" value="REGISTER"></div>
                </form>
                <form action="index.php" method="POST">
                    <div><input type="text" name="username" placeholder="USERNAME"></div>
                    <div><input type="password" name="password" placeholder="PASSWORD"></div>
                    <div><input type="submit" name="login" value="LOGIN"></div>
                </form>
            </div>
        <?php } ?>
    </div>
</body>
</html>
