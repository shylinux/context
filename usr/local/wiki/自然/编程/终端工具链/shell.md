## shell

- 官网：<https://www.gnu.org/software/bash/>
- 文档：<https://www.gnu.org/software/bash/manual/>
- 源码：<http://ftp.gnu.org/gnu/bash/bash-4.3.30.tar.gz>



## 文件管理
```
etc lib dev usr
boot proc
root home
sbin bin
var tmp run sys
opt srv mnt media
```

ls
cp
ln
mv
rm
cd
pwd
mkdir
rmdir

cat
more
less
head
tail
stat
file
hexdump
objdump
touch

## 进程管理
ps
top
kill
killall
nice
renice

## 磁盘管理
df
du
mount
umount

find
grep
sort
tar

## 网络管理

## 权限管理
/etc/passwd
/etc/shadow
/etc/skel
useradd
userdel
usermod
passwd
chpasswd
chsh
chfn
chage
groupadd
groupmod
groupdel

umask
chmod
chown
chgrp

## 环境变量
env
set
unset
export
alias
PATH
PS1

## 使用命令
```
$=
<< < | > >>

nohup & C-C C-Z fg bg trap sleep jobs
crontab inittab rc.local
```

## 脚本编程
```
#! /bin/bash
echo "hello world"
```
source
return
bash
exit

```
[ -eq -gt -lt -ne -ge -le ]
[ < <= = >= > != -z -n ]
[ -d -f -e -r -w -x -nt -ot ]

if cmd; then cmd; elif cmd; then cmd; else cmd; fi
case var in cond) cmd;; esac

IFS= for var in list; do cmd; done
while true; do cmd; done
until false; do cmd; done
break continue

$0 $1 $# $* $@ shift
OPTARG= OPTINDEX= getopts fmt var
select var in list; do cmd; done
REPLY read

function local return
```



uname
umount
touch
tar
su
sleep
sed
rm
ps
ping
netstat
mv
nano
more
lsmod
ls
ln
less
kill
gzip
grep
false
echo
dd
dmesg
date
cp
chgrp
cat
bash
sh
chmod
chown
cpio
df
dir
gunzip
hostname
lsblk
mkdir
mknod
mount
pwd
rmdir
true
which


default
alternatives
adduser.conf

bash.bashrc

dhcp
timezone

init
init.d
systemd
rc.local
rc0.d
rc1.d
rc2.d
rc3.d
rc4.d
rc5.d
rc6.d
rcS.d

fstab
fstab.aliyun_backup
fstab.aliyun_backup.xen
fstab.d

ld.so.cache
ld.so.conf
ld.so.conf.d

passwd

inputrc
ssh
ssl
shells
skel
zsh

vim
pki
terminfo
zsh_command_not_found

network
networks
host.conf
hostname
hosts
hosts.allow
hosts.deny
qemu
qemu-ifdown
qemu-ifup

cron.d
cron.daily
cron.hourly
cron.monthly
cron.weekly
crontab

apt
perl
python
python2.7
python3
python3.4
rsyslog.conf
rsyslog.d
php5
X11
apache2
mysql

apm
apparmor
apparmor.d
backup
bash_completion
bash_completion.d
bindresvport.blacklist
blkid.conf
blkid.tab
ca-certificates
ca-certificates.conf
ca-certificates.conf.dpkg-old
calendar
chatscripts
cloud
console-setup
dbus-1
debconf.conf
debian_version
deluser.conf
depmod.d
dictionaries-common
discover-modprobe.conf
discover.conf.d
dpkg
drirc
emacs
environment
fonts
fuse.conf
gai.conf
gdb
groff
group
group-
grub.d
gshadow
gshadow-
hdparm.conf
initramfs-tools
insserv
insserv.conf
insserv.conf.d
iproute2
iscsi
issue
issue.net
kbd
kernel
kernel-img.conf
ldap
legal
libaudit.conf
libnl-3
locale.alias
localtime
logcheck
login.defs
logrotate.conf
logrotate.d
lsb-release
ltrace.conf
lynx-cur
magic
magic.mime
mailcap
mailcap.order
manpath.config
mime.types
mke2fs.conf
modprobe.d
modules
motd
mtab
nanorc
newt
nscd.conf
nsswitch.conf
ntp.conf
ntp.conf.backup
opt
os-release
pam.conf
pam.d
passwd-
pm
popularity-contest.conf
ppp
profile
profile.d
protocols
pulse
resolv.conf
resolvconf
rmt
rpc
securetty
security
selinux
sensors.d
sensors3.conf
services
sgml
shadow
shadow-
subgid
subgid-
subuid
subuid-
sudoers
sudoers.d
sysctl.conf
sysctl.d
sysstat
ucf.conf
udev
ufw
update-manager
update-motd.d
updatedb.conf
upstart-xsessions
vtrgb
wgetrc
xml

bunzip2
busybox
bzcat
bzcmp
bzdiff
bzegrep
bzexe
bzfgrep
bzgrep
bzip2
bzip2recover
bzless
bzmore
chacl
chvt
dash
dbus-cleanup-sockets
dbus-daemon
dbus-uuidgen
dnsdomainname
domainname
dumpkeys
ed
egrep
fgconsole
fgrep
findmnt
fuser
fusermount
getfacl
gzexe
ip
kbd_mode
kmod
lessecho
lessfile
lesskey
lesspipe
loadkeys
login
loginctl
lowntfs-3g
mktemp
mountpoint
mt
mt-gnu
nc
nc.openbsd
nc.traditional
netcat
nisdomainname
ntfs-3g
ntfs-3g.probe
ntfs-3g.secaudit
ntfs-3g.usermap
ntfscat
ntfsck
ntfscluster
ntfscmp
ntfsdump_logfile
ntfsfix
ntfsinfo
ntfsls
ntfsmftalloc
ntfsmove
ntfstruncate
ntfswipe
open
openvt
pidof
ping6
plymouth
plymouth-upstart-bridge
rbash
readlink
red
rnano
run-parts
running-in-container
rzsh
setfacl
setfont
setupcon
sh.distrib
ss
static-sh
stty
sync
tailf
tempfile
udevadm
ulockmgr_server
uncompress
unicode_start
vdir
whiptail
ypdomainname
zcat
zcmp
zdiff
zegrep
zfgrep
zforce
zgrep
zless
zmore
znew
zsh
zsh5



fdisk
fsck
halt
ifconfig
ldconfig
lsmod
mkfs
modinfo
reboot
rmmod
route
shutdown
modprobe
ifdown
ifquery
ifup




MAKEDEV
acpi_available
agetty
apm_available
apparmor_parser
badblocks
biosdevname
blkid
blockdev
bridge
capsh
cfdisk
crda
ctrlaltdel
debugfs
depmod
dhclient
dhclient-script
discover
discover-modprobe
discover-pkginstall
dmsetup
dosfsck
dosfslabel
dumpe2fs
e2fsck
e2image
e2label
e2undo
fatlabel
findfs
fsck.cramfs
fsck.ext2
fsck.ext3
fsck.ext4
fsck.ext4dev
fsck.fat
fsck.minix
fsck.msdos
fsck.nfs
fsck.vfat
fsfreeze
fstab-decode
fstrim
fstrim-all
getcap
getpcaps
getty
hdparm
hwclock
init
initctl
insmod
installkernel
ip
ip6tables
ip6tables-apply
ip6tables-restore
ip6tables-save
ipmaddr
iptables
iptables-apply
iptables-restore
iptables-save
iptunnel
isosize
kbdrate
killall5
ldconfig.real
logsave
losetup
mii-tool
mkdosfs
mke2fs
mkfs.bfs
mkfs.cramfs
mkfs.ext2
mkfs.ext3
mkfs.ext4
mkfs.ext4dev
mkfs.fat
mkfs.minix
mkfs.msdos
mkfs.ntfs
mkfs.vfat
mkhomedir_helper
mkntfs
mkswap
mntctl
mount.fuse
mount.lowntfs-3g
mount.ntfs
mount.ntfs-3g
mountall
nameif
ntfsclone
ntfscp
ntfslabel
ntfsresize
ntfsundelete
on_ac_power
pam_tally
pam_tally2
parted
partprobe
pivot_root
plipconfig
plymouthd
poweroff
rarp
raw
regdbdump
reload
resize2fs
resolvconf
restart
rtacct
rtmon
runlevel
setcap
setvtrgb
sfdisk
shadowconfig
slattach
start
start-stop-daemon
startpar
startpar-upstart-inject
status
stop
sulogin
swaplabel
swapoff
swapon
switch_root
sysctl
tc
telinit
tune2fs
udevadm
udevd
unix_chkpwd
unix_update
upstart-dbus-bridge
upstart-event-bridge
upstart-file-bridge
upstart-local-bridge
upstart-socket-bridge
upstart-udev-bridge
ureadahead
wipefs
xtables-multi
