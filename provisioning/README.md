# provisioning

問題サーバの構築用ansible playbook群



## files

- `ansible.cfg`: basic configuration of ansible. not need to edit.
- `hosts`: hosts info. please edit dealing with your env.
- `ssh_config`: ditto
- `challenges.yaml`: configrutaion of challenges server. not need to edit as you use EC2 or your private project.
- `proxy.yaml`: configrutaion of proxy server. not need to edit as you use EC2 or your private project.
- `monitor_client.yml`: playbook to collect docker/node info to prometheus server.
- `monitor_server.yml`: playbook to deploy prometheus and grafana server. you MUST change the grafana admin's password.
- `deploy_challenge.yml`: playbook to deploy challenge to the server. no need to edit
- `apply_proxy.yml`: playbook to apply new proxy setting to the proxy server. no need to edit
- `ip_addrs.json`: ip address map from ssh target name to (private) ip address

- `templates/`: static files


## provisioning

```
$ ansible-playbook -i hosts proxy.yaml
$ ansible-playbook -i hosts challenges.yaml
```

if you want to use metrics of challenge servers

```
$ ansible-playbook -i hosts monitor_client.yml
$ ansible-playbook -i hosts monitor_server.yml
```

## deploy challenge

```
$ ./deploy -c ../challenges/just_login --host node-2 --proxy node-1
```

## todos

- [x] collect haproxy's metrics
