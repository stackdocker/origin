
## Prerequisite

[vagrant@localhost origin]$ ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default 
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether 08:00:27:24:23:96 brd ff:ff:ff:ff:ff:ff
    inet 10.0.2.15/24 brd 10.0.2.255 scope global dynamic eth0
       valid_lft 56943sec preferred_lft 56943sec
    inet6 fe80::a00:27ff:fe24:2396/64 scope link 
       valid_lft forever preferred_lft forever
3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether 08:00:27:8d:39:0a brd ff:ff:ff:ff:ff:ff
    inet 172.17.4.50/24 brd 172.17.4.255 scope global eth1
       valid_lft forever preferred_lft forever
    inet6 fe80::a00:27ff:fe8d:390a/64 scope link 
       valid_lft forever preferred_lft forever
4: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default 
    link/ether 02:42:4e:6f:f4:c1 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.1/22 scope global docker0
       valid_lft forever preferred_lft forever
    inet6 fe80::42:4eff:fe6f:f4c1/64 scope link 
       valid_lft forever preferred_lft forever
15: br-e0454ceaa2f8: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default 
    link/ether 02:42:47:7d:5a:82 brd ff:ff:ff:ff:ff:ff
    inet 10.1.0.1/24 scope global br-e0454ceaa2f8
       valid_lft forever preferred_lft forever

## Configure docker

[vagrant@localhost origin]$ sudo systemctl stop docker.service

[vagrant@localhost origin]$ cat /etc/sysconfig/docker
# /etc/sysconfig/docker

# Modify these options if you want to change the way the docker daemon runs
OPTIONS='--bip=172.17.0.1/22 --insecure-registry=172.30.0.0/16,10.3.0.0/24 --selinux-enabled --log-driver=journald'
DOCKER_CERT_PATH=/etc/docker

# Enable insecure registry communication by appending the registry URL
# to the INSECURE_REGISTRY variable below and uncommenting it
# INSECURE_REGISTRY='--insecure-registry '

# On SELinux System, if you remove the --selinux-enabled option, you
# also need to turn on the docker_transition_unconfined boolean.
# setsebool -P docker_transition_unconfined

# Location used for temporary files, such as those created by
# docker load and build operations. Default is /var/lib/docker/tmp
# Can be overriden by setting the following environment variable.
# DOCKER_TMPDIR=/var/tmp

# Controls the /etc/cron.daily/docker-logrotate cron job status.
# To disable, uncomment the line below.
# LOGROTATE=false

[vagrant@localhost origin]$ sudo systemctl start docker.service


[vagrant@localhost origin]$ docker network ls
NETWORK ID          NAME                DRIVER
8e25300fc6d5        bridge              bridge              
b3ca82dc23ce        none                null                
da69cf450770        host                host          

[vagrant@localhost origin]$ docker network create --gateway=10.1.0.1 --subnet=10.1.0.0/24 cbr0
e0454ceaa2f8a34a49e5b01bb763f04b3c9730c3827f51a9915af369235409d2

[vagrant@localhost origin]$ docker network ls
NETWORK ID          NAME                DRIVER
8e25300fc6d5        bridge              bridge              
b3ca82dc23ce        none                null                
da69cf450770        host                host                
e0454ceaa2f8        cbr0                bridge   

[vagrant@localhost origin]$ sudo systemctl -l status docker.service
● docker.service - Docker Application Container Engine
   Loaded: loaded (/etc/systemd/system/docker.service; enabled; vendor preset: disabled)
   Active: active (running) since 二 2016-05-10 22:32:54 UTC; 3min 28s ago
     Docs: http://docs.docker.com
 Main PID: 3553 (docker)
   CGroup: /system.slice/docker.service
           └─3553 /usr/bin/docker daemon --bip=172.17.0.1/22 --insecure-registry=172.30.0.0/16,10.3.0.0/24 --selinux-enabled --log-driver=journald -s devicemapper --storage-opt dm.datadev=/dev/vg_vagrant/docker-data --storage-opt dm.metadatadev=/dev/vg_vagrant/docker-metadata

5月 10 22:32:54 localhost docker[3553]: time="2016-05-10T22:32:54.091056103Z" level=info msg="Loading containers: start."
5月 10 22:32:54 localhost docker[3553]: ......................................
5月 10 22:32:54 localhost docker[3553]: time="2016-05-10T22:32:54.102026666Z" level=info msg="Loading containers: done."
5月 10 22:32:54 localhost docker[3553]: time="2016-05-10T22:32:54.102409951Z" level=info msg="Daemon has completed initialization"
5月 10 22:32:54 localhost docker[3553]: time="2016-05-10T22:32:54.102742853Z" level=info msg="Docker daemon" commit=7206621 execdriver=native-0.2 graphdriver=devicemapper version=1.9.1
5月 10 22:32:54 localhost systemd[1]: Started Docker Application Container Engine.
5月 10 22:32:54 localhost docker[3553]: time="2016-05-10T22:32:54.115651591Z" level=info msg="API listen on /var/run/docker.sock"
5月 10 22:32:58 localhost docker[3553]: time="2016-05-10T22:32:58.673707844Z" level=info msg="{Action=networks, Username=vagrant, LoginUID=1000, PID=3615}"
5月 10 22:35:02 localhost docker[3553]: time="2016-05-10T22:35:02.873916319Z" level=info msg="{Action=create, Username=vagrant, LoginUID=1000, PID=3620}"
5月 10 22:35:06 localhost docker[3553]: time="2016-05-10T22:35:06.959990708Z" level=info msg="{Action=networks, Username=vagrant, LoginUID=1000, PID=3685}"

[vagrant@localhost origin]$ sudo journalctl -u docker.service -e
...
5月 10 22:32:54 localhost systemd[1]: Started Docker Application Container Engine.
5月 10 22:32:54 localhost docker[3553]: time="2016-05-10T22:32:54.115651591Z" level=info msg="API listen on /var/run/docker.sock"
5月 10 22:32:58 localhost docker[3553]: time="2016-05-10T22:32:58.673707844Z" level=info msg="{Action=networks, Username=vagrant,
5月 10 22:35:02 localhost docker[3553]: time="2016-05-10T22:35:02.873916319Z" level=info msg="{Action=create, Username=vagrant, L
5月 10 22:35:06 localhost docker[3553]: time="2016-05-10T22:35:06.959990708Z" level=info msg="{Action=networks, Username=vagrant,

### Docker images

[vagrant@localhost origin]$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES

[vagrant@localhost origin]$ docker images
REPOSITORY                                       TAG                  IMAGE ID            CREATED             VIRTUAL SIZE
tangfeixiong/openshift-origin                    latest               e21f3ead5f3c        2 hours ago         473.8 MB
gcr.io/google_containers/hyperkube-amd64         v1.2.4               b65f775dbf89        4 days ago          316.7 MB
gcr.io/google_containers/hyperkube-amd64         v1.3.0-alpha.2       c11527c21c02        4 weeks ago         398.5 MB
gcr.io/google_containers/exechealthz             1.0                  d6fccb55b399        5 weeks ago         7.116 MB
quay.io/tangfeixiong/netcat-http-server-simple   latest               7fa32f504c61        7 weeks ago         7.807 MB
gcr.io/google_containers/kube2sky                1.14                 c0b611ff3f70        9 weeks ago         27.8 MB
openshift/origin-release                         latest               6791072422b1        9 weeks ago         715.2 MB
openshift/origin-haproxy-router-base             latest               a0328f433acf        9 weeks ago         290.9 MB
openshift/origin-base                            latest               f2ffca9a8520        9 weeks ago         271.8 MB
docker.io/centos                                 centos7              2933d50b9f77        11 weeks ago        196.6 MB
gcr.io/google_containers/etcd-amd64              2.2.1                202873aab189        3 months ago        28.19 MB
gcr.io/google_containers/skydns                  2015-10-13-8c72f8c   d8ed451aa9b9        7 months ago        40.55 MB
gcr.io/google_containers/pause                   2.0                  8950680a606c        7 months ago        350.2 kB
docker.io/openshift/etcd-20-centos7              latest               7857141e9bb1        10 months ago       244.3 MB


## Configure Etcd2

[vagrant@localhost origin]$ sudo systemctl start etcd2.service

[vagrant@localhost origin]$ sudo systemctl -l status etcd2.service
● etcd2.service - etcd2
   Loaded: loaded (/etc/systemd/system/etcd2.service; disabled; vendor preset: disabled)
   Active: active (running) since 二 2016-05-10 22:42:55 UTC; 15s ago
 Main PID: 3725 (etcd)
   CGroup: /system.slice/etcd2.service
           └─3725 /data/bin/etcd

5月 10 22:42:55 localhost etcd[3725]: added member ce2a822cea30bfca [http://localhost:2380 http://localhost:7001] to cluster 7e27652122e8b2ae from store
5月 10 22:42:55 localhost etcd[3725]: set the cluster version to 2.3 from store
5月 10 22:42:55 localhost etcd[3725]: starting server... [version: 2.3.0, cluster version: 2.3]
5月 10 22:42:55 localhost systemd[1]: Started etcd2.
5月 10 22:42:55 localhost etcd[3725]: ce2a822cea30bfca is starting a new election at term 2
5月 10 22:42:55 localhost etcd[3725]: ce2a822cea30bfca became candidate at term 3
5月 10 22:42:55 localhost etcd[3725]: ce2a822cea30bfca received vote from ce2a822cea30bfca at term 3
5月 10 22:42:55 localhost etcd[3725]: ce2a822cea30bfca became leader at term 3
5月 10 22:42:55 localhost etcd[3725]: raft.node: ce2a822cea30bfca elected leader ce2a822cea30bfca at term 3
5月 10 22:42:55 localhost etcd[3725]: published {Name:default ClientURLs:[http://localhost:2379 http://localhost:4001]} to cluster 7e27652122e8b2ae

[vagrant@localhost origin]$ sudo ls /var/lib/etcd2/member
snap  wal

## Configure kubelet

[vagrant@localhost origin]$ sudo systemctl start kubelet.service

[vagrant@localhost origin]$ sudo journalctl -e -u docker.service --no-pager
...
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.296970       1 handlers.go:152] GET /api/v1/pods?fieldSelector=spec.nodeNa
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.298324       1 handlers.go:152] GET /api/v1/nodes?fieldSelector=spec.unsch
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.298899       1 handlers.go:152] GET /api/v1/persistentvolumes?resourceVers
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.300232       1 handlers.go:152] GET /api/v1/persistentvolumeclaims?resourc
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.300932       1 handlers.go:152] GET /api/v1/services?resourceVersion=0: (6
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.301562       1 handlers.go:152] GET /api/v1/replicationcontrollers?resourc
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.302679       1 handlers.go:152] GET /apis/extensions/v1beta1/replicasets?r
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.304152       1 handlers.go:152] GET /api/v1/pods?fieldSelector=spec.nodeNa
5月 10 22:51:22 localhost docker[3553]: I0510 22:51:22.851642       1 handlers.go:152] GET /api/v1/namespaces/kube-system/pods/ku

[vagrant@localhost origin]$ sudo journalctl -f -u kubelet.service
...
5月 10 23:04:12 localhost kubelet[10389]: I0510 23:04:12.432714   10389 kubelet.go:2420] SyncLoop (PLEG): "kube-controller-manager-172.17.4.50_kube-system(8de7ed5077fb91868b9694ccd2c26c02)", event: &pleg.PodLifecycleEvent{ID:"8de7ed5077fb91868b9694ccd2c26c02", Type:"ContainerStarted", Data:"d175a56f5b42af28402057f1b728ed4a7483968c3f21e15a0cac4aa7ebb20227"}
5月 10 23:04:12 localhost kubelet[10389]: E0510 23:04:12.441129   10389 kubelet.go:1767] Failed creating a mirror pod for "kube-scheduler-172.17.4.50_kube-system(e04310159b684c5a4d153aafc8acb114)": namespaces "kube-system" not found
5月 10 23:04:12 localhost kubelet[10389]: E0510 23:04:12.444163   10389 kubelet.go:1767] Failed creating a mirror pod for "kube-controller-manager-172.17.4.50_kube-system(8de7ed5077fb91868b9694ccd2c26c02)": namespaces "kube-system" not found
5月 10 23:04:13 localhost kubelet[10389]: E0510 23:04:13.440711   10389 kubelet.go:1767] Failed creating a mirror pod for "kube-controller-manager-172.17.4.50_kube-system(8de7ed5077fb91868b9694ccd2c26c02)": namespaces "kube-system" not found


[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig get nodes
NAME          STATUS    AGE
172.17.4.50   Ready     1m

[vagrant@localhost origin]$ docker ps
CONTAINER ID        IMAGE                                                     COMMAND                  CREATED              STATUS              PORTS               NAMES
990b08b97b94        gcr.io/google_containers/hyperkube-amd64:v1.2.4           "/hyperkube controlle"   17 seconds ago       Up 15 seconds                           k8s_kube-controller-manager.576108e7_kube-controller-manager-172.17.4.50_kube-system_8de7ed5077fb91868b9694ccd2c26c02_72d01df3
1d7e3516fa34        gcr.io/google_containers/hyperkube-amd64:v1.2.4           "/hyperkube scheduler"   17 seconds ago       Up 16 seconds                           k8s_kube-scheduler.837c996c_kube-scheduler-172.17.4.50_kube-system_e04310159b684c5a4d153aafc8acb114_12f060eb
ce0cb5d4e65a        gcr.io/google_containers/hyperkube-amd64:v1.3.0-alpha.2   "/hyperkube proxy --m"   About a minute ago   Up About a minute                       k8s_kube-proxy.d7767979_kube-proxy-172.17.4.50_kube-system_bc97a54621f44c5ad5aa9d894e701a86_c5e47a23
e5c5d26769a8        gcr.io/google_containers/hyperkube-amd64:v1.2.4           "/hyperkube apiserver"   About a minute ago   Up About a minute                       k8s_kube-apiserver.d5a06cbd_kube-apiserver-172.17.4.50_kube-system_ccf0d7c54a8585bbc94dcbf2c4a3cafe_cd7bab18
e9e8db9c52d9        gcr.io/google_containers/pause:2.0                        "/pause"                 About a minute ago   Up About a minute                       k8s_POD.6059dfa2_kube-proxy-172.17.4.50_kube-system_bc97a54621f44c5ad5aa9d894e701a86_ee79e3be
23fa086d9da2        gcr.io/google_containers/pause:2.0                        "/pause"                 About a minute ago   Up About a minute                       k8s_POD.6059dfa2_kube-apiserver-172.17.4.50_kube-system_ccf0d7c54a8585bbc94dcbf2c4a3cafe_6586f800
a0bfc5c793a9        gcr.io/google_containers/pause:2.0                        "/pause"                 About a minute ago   Up About a minute                       k8s_POD.6059dfa2_kube-scheduler-172.17.4.50_kube-system_e04310159b684c5a4d153aafc8acb114_60f3fa04
33f930449104        gcr.io/google_containers/pause:2.0                        "/pause"                 About a minute ago   Up About a minute                       k8s_POD.6059dfa2_kube-controller-manager-172.17.4.50_kube-system_8de7ed5077fb91868b9694ccd2c26c02_be931f59

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig create -f etc/kubernetes/addons/kube-system.json 
namespace "kube-system" created

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig create -f etc/kubernetes/addons/kube-dns-rc.json 

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig create -f etc/kubernetes/addons/kube-dns-svc.json 


[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig get pods --all-namespaces
NAMESPACE     NAME                                  READY     STATUS             RESTARTS   AGE
kube-system   kube-apiserver-172.17.4.50            1/1       Running            0          11m
kube-system   kube-controller-manager-172.17.4.50   1/1       Running            0          44s
kube-system   kube-dns-v9-3453v                     2/4       CrashLoopBackOff   9          9m
kube-system   kube-proxy-172.17.4.50                1/1       Running            0          11m
kube-system   kube-scheduler-172.17.4.50            1/1       Running            0          45s


### Test

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig run nc-http --image=quay.io/tangfeixiong/netcat-http-server-simple --port=80 --expose=true
service "nc-http" created
deployment "nc-http" created

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig get service
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   10.3.0.1     <none>        443/TCP   17m
nc-http      10.3.0.15    <none>        80/TCP    32s

[vagrant@localhost origin]$ curl 10.3.0.15
<html>
        <head>
                <title>Hello Page</title>
        </head>
        <body>
                <h1>Hello</h1>
                <h2>Container</h2>
                <p>Powered by nc</p>
        </body>
</html>

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig get ep
NAME         ENDPOINTS         AGE
kubernetes   172.17.4.50:443   18m
nc-http      172.17.0.3:80     1m

[vagrant@localhost origin]$ curl 172.17.0.3
<html>
        <head>
                <title>Hello Page</title>
        </head>
        <body>
                <h1>Hello</h1>
                <h2>Container</h2>
                <p>Powered by nc</p>
        </body>
</html>

