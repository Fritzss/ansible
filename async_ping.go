package main

import (
        "context"
        "encoding/csv"
        "fmt"
        "golang.org/x/net/icmp"
        "golang.org/x/net/ipv4"
        "net"
        "os"
        "strconv"
        "sync"
        "time"
)

// ICMP payload template
func icmpEchoRequest(id int, seq int) []byte {
        // Build a standard ICMP Echo request payload
        // Type 8 (Echo), Code 0
        // Identifier and Sequence in the payload (we'll include 2xuint16)
        // We'll keep the body simple: "PING"
        msg := icmp.Message{
                Type: ipv4.ICMPTypeEcho,
                Code: 0,
                Body: &icmp.Echo{
                        ID:   id,
                        Seq:  seq,
                        Data: []byte("PING"),
                },
        }
        b, _ := msg.Marshal(nil)
        return b
}

// send a single ICMP echo to target using a raw socket (requires root/admin)
func sendICMPEcho(ctx context.Context, dst string, id int, seq int, timeout time.Duration) (bool, time.Duration, error) {
        start := time.Now()

        // Resolve address
        raddr, err := net.ResolveIPAddr("ip4", dst)
        if err != nil {
                return false, 0, err
        }

        // Open raw socket
        c, err := net.DialIP("ip4:icmp", nil, raddr)
        if err != nil {
                return false, 0, err
        }
        defer c.Close()

        // Create ICMP Echo request
        b := icmpEchoRequest(id, seq)

        // Set deadline
        _ = c.SetDeadline(time.Now().Add(timeout))

        // Send
        _, err = c.Write(b)
        if err != nil {
                return false, 0, err
        }

        // Prepare to read reply
        // ICMP reply header + body
        reply := make([]byte, 1500)

        _, err = c.Read(reply)
        elapsed := time.Since(start)

        // Basic success if we got a reply before timeout
        if err != nil {
                return false, elapsed, err
        }
        // Parse reply (optional: ensure it's Echo Reply and id/seq match)
        rm, err := icmp.ParseMessage(1, reply)
        if err != nil {
                return false, elapsed, err
        }
        switch rm.Type {
        case ipv4.ICMPTypeEchoReply:
                return true, elapsed, nil
        default:
                // other ICMP types; still consider as failed to match EchoReply
                return false, elapsed, nil
        }
}

// worker polls a given subnet and checks each IP asynchronously
func worker(ctx context.Context, subnet *net.IPNet, startID int, timeout time.Duration, delay time.Duration, results chan<- []string, wg *sync.WaitGroup) {
        defer wg.Done()
        // iterate all hosts in subnet
        // We'll use a simple approach: enumerate all host addresses in network (excluding network and broadcast)
        ones, bits := subnet.Mask.Size()
        // Ensure IPv4
        if ones == 0 && bits == 0 {
                return
        }
        network := subnet.IP.To4()
         if network == nil {
                 return
         }
        // Compute start/end
        // Convert to uint32
        ni := ip4ToUint32(network)
        mask := ip4ToUint32(net.IP(subnet.Mask).To4())
        networkBase := ni & mask
        broadcast := networkBase | (^mask)

        id := startID

        // Iterate usable addresses (excluding network and broadcast)
        for ip := networkBase + 1; ip < broadcast; ip++ {
                target := uint32ToIPv4(ip).String()
                // skip if the string equals network or broadcast already handled
                // spawn goroutine per host with small limit
                select {
                case <-ctx.Done():
                        return
                default:
                }
                // Launch a goroutine per address but throttle using a small Worker pool controlled by delay
                go func(dst string, iid int) {
                        // each attempt uses a timeout
                        ok, dur, err := sendICMPEcho(ctx, dst, iid, 1, timeout)
                        res := []string{dst, strconv.FormatBool(ok), dur.String()}
                        if err != nil {
                                // record error as third field if needed
                                res[2] = dur.String() + " (" + err.Error() + ")"
                        }
                        // send result
                        select {
                        case <-ctx.Done():
                                return
                        case results <- res:
                        }
                }(target, id)
                id++
                // small delay to avoid rush
                time.Sleep(delay)
        }
}

// helpers for IP<->uint32 conversions
func ip4ToUint32(ip net.IP) uint32 {
        return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
func uint32ToIPv4(n uint32) net.IP {
        return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func main() {
        // Example usage parameters (adjust as needed)
        // Subnet in CIDR notation, e.g. 192.168.1.0/24
        var subnet string
        fmt.Scanln(&subnet)
        fmt.Print(subnet)
        subnetCIDR := subnet
        // "192.168.1.0/24"
        // timeout for each ICMP echo
        timeout := 250 * time.Millisecond
        // delay between launching each probe (to throttle)
        delay := 5 * time.Millisecond
        // total duration to run (we won't rely on it; we'll cancel via context)
        // Output CSV: address, success(true/false), latency
        // or stdout with header

        _, ipnet, err := net.ParseCIDR(subnetCIDR)
        if err != nil {
                fmt.Fprintf(os.Stderr, "invalid CIDR: %v\n", err)
                return
        }

        // Prepare output
        out := csv.NewWriter(os.Stdout)
        defer out.Flush()
        // header
        _ = out.Write([]string{"ip", "reachable", "latency"})

        // Context to cancel
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        // Channel for results
        results := make(chan []string)

        var wg sync.WaitGroup
        wg.Add(1)
        // start worker to enqueue probes
        go worker(ctx, ipnet, 1000, timeout, delay, results, &wg)

        // Collect results for a bounded time or until all goroutines finish
        // We'll stop after a fixed duration (e.g., 30 seconds) or when all IPs are processed.
        collectDone := make(chan struct{})
        go func() {
                wg.Wait()
                close(collectDone)
        }()

        // Consume results and write to CSV
        for {
                select {
                case r := <-results:
                        _ = out.Write(r)
                case <-collectDone:
                        return
                }
        }
}
