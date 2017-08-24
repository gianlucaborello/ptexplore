# ptexplore (page table explorer)

A simple tool to print the page table content for a process in Linux.

Works by attaching to a process and printing each memory area. Optionally, memory areas can be restricted via a filter.

## Install

```
$ go install github.com/gianlucaborello/ptexplore/cmd/ptexplore
```

## Usage

```
$ ptexplore --help
Explore page table of a process under Linux.
Works by attaching to a process and printing each memory area. Optionally, memory areas can be restricted via a filter.

Usage of ptexplore:
  -address string
        Analyze a single address (e.g. '0x7f66a002ab70')
  -areas string
        Comma separated list of memory areas (even patterns) to analyze (e.g. 'stack,heap,libc')
  -pid int
        Pid of the process to analyze (e.g. 42)
  -quiet
        Don't print page table details, just a summary of the memory areas
```

Print a summary of all the memory areas (similar to `/proc/PID/maps`):

```
$ ptexplore -pid 121361 --quiet
Area '/usr/bin/cadvisor', range 0x0000000000400000 - 0x00000000010a7000 (13 MB)
Area '/usr/bin/cadvisor', range 0x00000000012a6000 - 0x0000000001343000 (643 kB)
Area 'anonymous', range 0x0000000001343000 - 0x0000000001370000 (184 kB)
Area '[heap]', range 0x000000000290c000 - 0x000000000292f000 (143 kB)
Area 'anonymous', range 0x000000c000000000 - 0x000000c000005000 (20 kB)
Area 'anonymous', range 0x000000c41ff68000 - 0x000000c421200000 (20 MB)
Area '/usr/glibc-compat/lib/libc-2.23.so', range 0x00007f669e3bc000 - 0x00007f669e3c0000 (16 kB)
Area '/usr/glibc-compat/lib/libc-2.23.so', range 0x00007f669e3c0000 - 0x00007f669e3c2000 (8.2 kB)
Area '[stack]', range 0x00007fff5e57d000 - 0x00007fff5e59e000 (135 kB)
Area '[vvar]', range 0x00007fff5e5e4000 - 0x00007fff5e5e7000 (12 kB)
Area '[vdso]', range 0x00007fff5e5e7000 - 0x00007fff5e5e9000 (8.2 kB)
```

Print a filtered summary of all the memory areas:

```
$ ptexplore -pid 121361 --areas stack,heap --quiet
Area '[stack]', range 0x00007fff5e57d000 - 0x00007fff5e59e000 (135 kB)
Area '[heap]', range 0x000000000290c000 - 0x000000000292f000 (143 kB)
```

Print a filtered view of the page table by memory area. For every memory areas, present and swapped pages are shown, along with their count, physical address and flags:

```
$ ptexplore -pid 121361 --areas stack,heap

Area '[stack]', range 0x00007fff5e57d000 - 0x00007fff5e59e000 (135 kB)

... 30 non mapped pages ...
0x00007fff5e59b000: physical address: 0x000000022762a000 exclusive soft-dirty count:1 flags:UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x00007fff5e59c000: physical address: 0x0000000221c39000 exclusive soft-dirty count:1 flags:UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x00007fff5e59d000: physical address: 0x00000001a3f4b000 exclusive soft-dirty count:1 flags:REFERENCED,UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED

Area '[heap]', range 0x000000000290c000 - 0x000000000292f000 (143 kB)

0x000000000290c000: physical address: 0x00000002282fe000 exclusive soft-dirty count:1 flags:UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x000000000290d000: physical address: 0x000000022bbae000 exclusive soft-dirty count:1 flags:REFERENCED,UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x000000000290e000: physical address: 0x000000022bbaf000 exclusive soft-dirty count:1 flags:UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x000000000290f000: physical address: 0x0000000199fce000 exclusive soft-dirty count:1 flags:REFERENCED,UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x0000000002910000: physical address: 0x00000001aa074000 exclusive soft-dirty count:1 flags:REFERENCED,UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
0x0000000002911000: physical address: 0x000000022006e000 exclusive soft-dirty count:1 flags:REFERENCED,UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
... 29 non mapped pages ...
```

Print the details for a specific page given a virtual memory address:

```
$ ptexplore --pid 121361 --address 0x7f66a002ab70 

Area 'anonymous', range 0x00007f669f82c000 - 0x00007f66a002c000 (8.4 MB)

0x00007f66a002a000: physical address: 0x000000019fe8c000 exclusive soft-dirty count:1 flags:UPTODATE,LRU,ACTIVE,MMAP,ANON,SWAPBACKED
```
