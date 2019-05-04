## linux

- 官网：<https://www.linux.org/>
- 文档：<https://www.kernel.org/doc/html/latest/>
- 源码：<https://mirrors.edge.kernel.org/pub/linux/kernel/v3.x/linux-3.7.4.tar.gz>
- 开源：<https://github.com/torvalds/linux>

- 代理：<https://mirror.tuna.tsinghua.edu.cn/kernel/v3.x/linux-3.7.4.tar.gz>

## 内核源码

lib
arch
init
kernel
include
drivers
ipc
net
mm
fs

firmware
security
crypto
sound
block
virt

tools
scripts
samples
Makefile
Kconfig
Kbuild
usr

README
COPYING
CREDITS
Documentation
MAINTAINERS
REPORTING-BUGS

## 系统启动
```
startup_32() // linux-2.6.12/arch/i386/kernel/head.S:57
start_kernel() // linux-2.6.12/init/main.c:424
	sched_init()
    trap_init() // arch/i386/kernel/traps.c
	init_IRQ()
	init_timers()
	softirq_init()
    time_init() // arch/i386/kernel/time.c
        time_init_hook()
            setup_irq(0, timer_interrupt)
```

## 进程调度
```
schedule()
```

## 系统中断
```
interrupt() // arch/i386/kernel/entry.S:397
do_IRQ() // arch/i386/kernel/irq.c:51
    __do_IRQ() // kernel/irq/handle.c:107
        handle_IRQ_event()

do_softirq() // arch/i386/kernel/softirq.c:158
    __do_softirq() // kernel/softirq.c:74

tasklet_schedule() // include/linux/interrupt.h
    __tasklet_schedule() // kernel/softirq.c:223

run_workqueue() // kernel/workqueue.c:145
```

## 并发同步
```
atomic_add() // include/asm-i386/atomic.h
spin_lock_irq() // include/linux/spinlock.h
down_interruptible() // include/asm-i386/semaphore.h
wait_for_completion() // kernel/sched.c
```

## 时间管理
```
timer_interrupt() // arch/i386/kernel/time.c:292
do_timer_interrupt() // arch/i386/kernel/time.c:250
    do_timer_interrupt_hook()
        do_timer() // kernel/timer.c:925
```

## 系统调用
```
system_call() // arch/i386/kernel/entry.S:225
sys_call_table[] // arch/i386/kernel/syscall_table.S
	.long sys_restart_syscall	/* 0 - old "setup()" system call, used for restarting */
	.long sys_exit() // kernel/exit.c:859
        do_exit()
        exit_mm()
        exit_sem()
        __exit_files()
        __exit_fs()
        exit_namespace()
        exit_thread()
        cpuset_exit()
        exit_keys()
        exit_notify()
        schedule()
	.long sys_fork // arch/i386/kernel/process.c:645
        do_fork() // kernel/fork.c:1192
            copy_process()
                dup_task_struct()
                copy_files()
                copy_fs()
                copy_sighand()
                copy_signal()
                copy_mm()
                sched_fork()
            wake_up_new_task()
	.long sys_read
	.long sys_write
	.long sys_open		/* 5 */
        get_unused_fd()
        filp_open()
            open_namei()
                path_lookup()
            dentry_open()
        fd_install()

	.long sys_close
	.long sys_waitpid
	.long sys_creat
	.long sys_link
	.long sys_unlink	/* 10 */
	.long sys_execve
	.long sys_chdir
	.long sys_time
...
```

## 进程管理
```
struct task_struct {
    // 进程状态
    state long / TASK_RUNNING // include/linux/sched.h:108

    // 线程信息
    thread_info *struct thread_info
        task *struct task_struct
        status

    // 进程树
    parent *struct task_struct
    children struct list_head
    sibling struct list_head

    mm *struct mm_struct
    binfmt *struct linux_binfmt
    thread struct thread_struct
        esp0
        eip
        esp
        fs
        gs

    fs *struct fs_struct
    files *struct files_struct
    signal *struct signal_struct

    pid pid_t
    tgit pid_t
}
```

## Compile
```
linux-3.7.4/Makefile
include $(srctree)/arch/$(SRCARCH)/Makefile # linux-3.7.4/Makefile:495
```
sudo apt-get install libncurses5-dev
make defconfig
make menuconfig
make
make tags

## GCC
gcc -E hi.c -o hi.i
gcc -S hi.c
gcc -c hi.c

ar -r hi.a hi.o he.o
gcc -shared -fPIC -o hi.so hi.c he.c
gcc -L./ -lhi main.c

nm
readelf -S hi.o
objdump -d hi.o
hexdump -C hi.o

## GRUB
- 下载：<ftp://ftp.gnu.org/gnu/grub/grub-2.00.tar.gz>
BIOS MBR GRUB

## 内核启动
```
startup_32() // arch/x86/kernel/head_32.S:88
i386_start_kernel() // arch/x86/kernel/head32.c:31
start_kernel() // init/main.c:468
    mm_init()
    sched_init()
    console_init()
    signals_init()
    rest_init()
        kernel_thread(kernel_init)
            do_fork() // kernel/fork.c:1548
                copy_process()
                wake_up_new_task() // kernel/sched/core.c:1622
                    activate_task()
                        enqueue_task(rq, p)

kernel_init() // init/main.c:805
    run_init_process()
        kernel_execve() // fs/exec.c:1710
            do_execve()
                do_execve_common()
                    open_exec()
                    sched_exec() // kernel/sched/core.c:2538
                        select_task_rq()
                    bprm_mm_init()
                    prepare_binprm()
                    search_binary_handler()
                        formats[i]->load_binary()

fair_sched_class() // kernel/sched/fair.c:5308
    select_task_rq_fair()
    pick_next_task_fair()
        pick_next_entity()
            __pick_first_entity()
                rb_entry()

formats[i]->load_binary()
    load_elf_binary() // fs/binfmt_elf.c:561
        current->mm->start_stack
        set_brk(elf_bss, elf_brk)

        start_thread() // arch/x86/kernel/process_32.c:200
formats[i]->load_binary()
    load_script()
        bprm_change_interp()
        open_exec()
        prepare_binprm()
        search_binary_handler()

struct task_struct { // linux-3.7.4/include/linux/sched.h:1190
    mm *struct mm_struct
        start_code
        end_code
        start_data
        end_data
        start_brk
        brk
        mmap_base
        start_stack
        arg_start
        arg_end
        env_start
        env_end

    thread struct thread_struct
        sp0
        sp
        es
        ds
        fsindex
        gsindex
        ip
        fs
        gs

    fs *struct fs_struct
    files *struct files_struct
    pid pid_t
    tgid pit_t
}
```
