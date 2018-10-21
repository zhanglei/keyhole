package keyhole

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PrintAllStats print all stats
func PrintAllStats(docs []ServerStatusDoc, span int) string {
	var lines []string
	lines = append(lines, printStatsDetails(docs, span))
	lines = append(lines, printGlobalLockDetails(docs, span))
	lines = append(lines, printLatencyDetails(docs, span))
	lines = append(lines, printMetricsDetails(docs, span))
	lines = append(lines, printWiredTigerCacheDetails(docs, span))
	lines = append(lines, printWiredTigerConcurrentTransactionsDetails(docs, span))
	return strings.Join(lines, "")
}

// printStatsDetails -
func printStatsDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	lines = append(lines, "\n--- Analytic Summary ---")
	lines = append(lines, "+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+")
	lines = append(lines, "| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |")
	lines = append(lines, "|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
		if cnt == 0 {
			stat1 = stat2
		} else if cnt == 1 {
			iops := stat2.OpCounters.Command - stat1.OpCounters.Command +
				stat2.OpCounters.Delete - stat1.OpCounters.Delete +
				stat2.OpCounters.Getmore - stat1.OpCounters.Getmore +
				stat2.OpCounters.Insert - stat1.OpCounters.Insert +
				stat2.OpCounters.Query - stat1.OpCounters.Query +
				stat2.OpCounters.Update - stat1.OpCounters.Update
			if d > 0 {
				iops = iops / d
			} else {
				iops = 0
			}
			if (stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults) >= 0 &&
				(stat2.OpCounters.Command-stat1.OpCounters.Command) >= 0 &&
				iops >= 0 {
				lines = append(lines, fmt.Sprintf("|%-25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.Mem.Resident,
					stat2.Mem.Virtual,
					stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults,
					stat2.OpCounters.Command-stat1.OpCounters.Command,
					stat2.OpCounters.Delete-stat1.OpCounters.Delete,
					stat2.OpCounters.Getmore-stat1.OpCounters.Getmore,
					stat2.OpCounters.Insert-stat1.OpCounters.Insert,
					stat2.OpCounters.Query-stat1.OpCounters.Query,
					stat2.OpCounters.Update-stat1.OpCounters.Update, iops))
			} else {
				cnt = 0
				lines = append(lines, "|-- REBOOT ---------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|")

			}
			stat1 = stat2
		} else if stat2.Host == stat1.Host {
			if cnt == len(docs)-1 || d >= span {
				iops := stat2.OpCounters.Command - stat1.OpCounters.Command +
					stat2.OpCounters.Delete - stat1.OpCounters.Delete +
					stat2.OpCounters.Getmore - stat1.OpCounters.Getmore +
					stat2.OpCounters.Insert - stat1.OpCounters.Insert +
					stat2.OpCounters.Query - stat1.OpCounters.Query +
					stat2.OpCounters.Update - stat1.OpCounters.Update
				if d > 0 {
					iops = iops / d
				} else {
					iops = 0
				}

				if (stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults) >= 0 &&
					(stat2.OpCounters.Command-stat1.OpCounters.Command) >= 0 &&
					iops >= 0 {
					lines = append(lines, fmt.Sprintf("|%-25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339),
						stat2.Mem.Resident,
						stat2.Mem.Virtual,
						stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults,
						stat2.OpCounters.Command-stat1.OpCounters.Command,
						stat2.OpCounters.Delete-stat1.OpCounters.Delete,
						stat2.OpCounters.Getmore-stat1.OpCounters.Getmore,
						stat2.OpCounters.Insert-stat1.OpCounters.Insert,
						stat2.OpCounters.Query-stat1.OpCounters.Query,
						stat2.OpCounters.Update-stat1.OpCounters.Update, iops))
				} else {
					cnt = 0
					lines = append(lines, "|-- REBOOT ---------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|")
				}
				stat1 = stat2
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+")
	return strings.Join(lines, "\n")
}

// printLatencyDetails -
func printLatencyDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	lines = append(lines, "\n--- Latencies Summary (ms) ---")
	lines = append(lines, "+-------------------------+----------+----------+----------+")
	lines = append(lines, "| Date/Time               | reads    | writes   | commands |")
	lines = append(lines, "|-------------------------|----------|----------|----------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt == 0 {
			stat1 = stat2
		} else if cnt == 1 {
			r := stat2.OpLatencies.Reads.Ops - stat1.OpLatencies.Reads.Ops
			if r > 0 {
				r = (stat2.OpLatencies.Reads.Latency - stat1.OpLatencies.Reads.Latency) / r
			}
			w := stat2.OpLatencies.Writes.Ops - stat1.OpLatencies.Writes.Ops
			if w > 0 {
				w = (stat2.OpLatencies.Writes.Latency - stat1.OpLatencies.Writes.Latency) / w
			}
			c := stat2.OpLatencies.Commands.Ops - stat1.OpLatencies.Commands.Ops
			if c > 0 {
				c = (stat2.OpLatencies.Commands.Latency - stat1.OpLatencies.Commands.Latency) / c
			}
			if r >= 0 && w >= 0 && c >= 0 {
				lines = append(lines, fmt.Sprintf("|%-25s|%10d|%10d|%10d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339), r/1000, w/1000, c/1000))
			} else {
				cnt = 0
				lines = append(lines, "|-- REBOOT ---------------|----------|----------|----------|")
			}
			stat1 = stat2
		} else if stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
				r := stat2.OpLatencies.Reads.Ops - stat1.OpLatencies.Reads.Ops
				if r > 0 {
					r = (stat2.OpLatencies.Reads.Latency - stat1.OpLatencies.Reads.Latency) / r
				}
				w := stat2.OpLatencies.Writes.Ops - stat1.OpLatencies.Writes.Ops
				if w > 0 {
					w = (stat2.OpLatencies.Writes.Latency - stat1.OpLatencies.Writes.Latency) / w
				}
				c := stat2.OpLatencies.Commands.Ops - stat1.OpLatencies.Commands.Ops
				if c > 0 {
					c = (stat2.OpLatencies.Commands.Latency - stat1.OpLatencies.Commands.Latency) / c
				}
				if r >= 0 && w >= 0 && c >= 0 {
					lines = append(lines, fmt.Sprintf("|%-25s|%10d|%10d|%10d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339), r/1000, w/1000, c/1000))
				} else {
					cnt = 0
					lines = append(lines, "|-- REBOOT ---------------|----------|----------|----------|")
				}
				stat1 = stat2
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+----------+----------+----------+")
	return strings.Join(lines, "\n")
}

// printMetricsDetails -
func printMetricsDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	lines = append(lines, "\n--- Metrics ---")
	lines = append(lines, "+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+")
	lines = append(lines, "| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |")
	lines = append(lines, "|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt == 0 {
			stat1 = stat2
		} else if cnt == 1 {
			if (stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned) >= 0 &&
				(stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects) >= 0 &&
				(stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts) >= 0 &&
				(stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted) >= 0 &&
				(stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned) >= 0 {
				lines = append(lines, fmt.Sprintf("|%-25s|%10d|%12d|%12d|%14d|%10d|%10d|%10d|%10d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned,
					stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects,
					stat2.Metrics.Operation.ScanAndOrder-stat1.Metrics.Operation.ScanAndOrder,
					stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts,
					stat2.Metrics.Document.Deleted-stat1.Metrics.Document.Deleted,
					stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted,
					stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned,
					stat2.Metrics.Document.Updated-stat1.Metrics.Document.Updated))
			} else {
				cnt = 0
				lines = append(lines, "|-- REBOOT ---------------|----------|------------|------------|--------------|----------|----------|----------|----------|")
			}
			stat1 = stat2
		} else if stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
				if (stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned) >= 0 &&
					(stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects) >= 0 &&
					(stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts) >= 0 &&
					(stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted) >= 0 &&
					(stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned) >= 0 {
					lines = append(lines, fmt.Sprintf("|%-25s|%10d|%12d|%12d|%14d|%10d|%10d|%10d|%10d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339),
						stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned,
						stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects,
						stat2.Metrics.Operation.ScanAndOrder-stat1.Metrics.Operation.ScanAndOrder,
						stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts,
						stat2.Metrics.Document.Deleted-stat1.Metrics.Document.Deleted,
						stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted,
						stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned,
						stat2.Metrics.Document.Updated-stat1.Metrics.Document.Updated))
				} else {
					cnt = 0
					lines = append(lines, "|-- REBOOT ---------------|----------|------------|------------|--------------|----------|----------|----------|----------|")
				}
				stat1 = stat2
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+")
	return strings.Join(lines, "\n")
}

// printGlobalLockDetails prints globalLock stats
func printGlobalLockDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat := ServerStatusDoc{}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	acm := 0
	lines = append(lines, "\n--- Global Locks Summary ---")
	lines = append(lines, "+-------------------------+--------------+--------------------------------------------+--------------------------------------------+")
	lines = append(lines, "|                         | Total Time   | Active Clients                             | Current Queue                              |")
	lines = append(lines, "| Date/Time               | (ms)         | total        | readers      | writers      | total        | readers      | writers      |")
	lines = append(lines, "|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if cnt == 0 {
			stat1 = stat
			stat2.Host = stat1.Host
		} else if cnt == 1 {
			if (stat.GlobalLock.TotalTime - stat1.GlobalLock.TotalTime) >= 0 {
				lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|",
					stat.LocalTime.In(loc).Format(time.RFC3339),
					(stat.GlobalLock.TotalTime-stat1.GlobalLock.TotalTime)/1000,
					stat.GlobalLock.CurrentQueue.Total,
					stat.GlobalLock.CurrentQueue.Readers,
					stat.GlobalLock.CurrentQueue.Writers,
					stat.GlobalLock.CurrentQueue.Total,
					stat.GlobalLock.CurrentQueue.Readers,
					stat.GlobalLock.CurrentQueue.Writers))
			} else {
				cnt = 0
				lines = append(lines, "|-- REBOOT ---------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
			}
			stat1 = stat
			stat2 = ServerStatusDoc{}
			stat2.Host = stat1.Host
		} else if stat2.Host == stat.Host {
			d := int(stat.LocalTime.Sub(stat1.LocalTime).Seconds())
			acm++
			stat2.LocalTime = stat.LocalTime
			stat2.GlobalLock.TotalTime = stat.GlobalLock.TotalTime
			stat2.GlobalLock.CurrentQueue.Total = stat.GlobalLock.CurrentQueue.Total
			stat2.GlobalLock.CurrentQueue.Readers += stat.GlobalLock.CurrentQueue.Readers
			stat2.GlobalLock.CurrentQueue.Writers += stat.GlobalLock.CurrentQueue.Writers
			stat2.GlobalLock.CurrentQueue.Total += stat.GlobalLock.CurrentQueue.Total
			stat2.GlobalLock.CurrentQueue.Readers += stat.GlobalLock.CurrentQueue.Readers
			stat2.GlobalLock.CurrentQueue.Writers += stat.GlobalLock.CurrentQueue.Writers
			if cnt == len(docs)-1 || d >= span {
				if (stat.GlobalLock.TotalTime - stat1.GlobalLock.TotalTime) >= 0 {
					lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339),
						(stat2.GlobalLock.TotalTime-stat1.GlobalLock.TotalTime)/1000,
						stat2.GlobalLock.CurrentQueue.Total/acm,
						stat2.GlobalLock.CurrentQueue.Readers/acm,
						stat2.GlobalLock.CurrentQueue.Writers/acm,
						stat2.GlobalLock.CurrentQueue.Total/acm,
						stat2.GlobalLock.CurrentQueue.Readers/acm,
						stat2.GlobalLock.CurrentQueue.Writers/acm))
				} else {
					cnt = 0
					lines = append(lines, "|-- REBOOT ---------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
				}
				acm = 0
				stat1 = stat2
				stat2 = ServerStatusDoc{}
				stat2.Host = stat1.Host
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	return strings.Join(lines, "\n")
}

// printWiredTigerCacheDetails prints wiredTiger cache stats
func printWiredTigerCacheDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat := ServerStatusDoc{}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	acm := 0
	lines = append(lines, "\n--- WiredTiger Cache Summary ---")
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	lines = append(lines, "|                         | MaxBytes     | Currently    | Tracked      | Modified     | Unmodified   | PagesRead    | PagesWritten |")
	lines = append(lines, "| Date/Time               | Configured   | InCache      | DirtyBytes   | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |")
	lines = append(lines, "|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if cnt == 0 {
			stat1 = stat
			stat2.Host = stat1.Host
		} else if stat2.Host == stat.Host {
			d := int(stat.LocalTime.Sub(stat1.LocalTime).Seconds())
			acm++
			stat2.LocalTime = stat.LocalTime
			stat2.WiredTiger.Cache.MaxBytesConfigured += stat.WiredTiger.Cache.MaxBytesConfigured
			stat2.WiredTiger.Cache.CurrentlyInCache += stat.WiredTiger.Cache.CurrentlyInCache
			stat2.WiredTiger.Cache.TrackedDirtyBytes += stat.WiredTiger.Cache.TrackedDirtyBytes
			stat2.WiredTiger.Cache.ModifiedPagesEvicted += stat.WiredTiger.Cache.ModifiedPagesEvicted
			stat2.WiredTiger.Cache.UnmodifiedPagesEvicted += stat.WiredTiger.Cache.UnmodifiedPagesEvicted
			stat2.WiredTiger.Cache.PagesReadIntoCache += stat.WiredTiger.Cache.PagesReadIntoCache
			stat2.WiredTiger.Cache.PagesWrittenFromCache += stat.WiredTiger.Cache.PagesWrittenFromCache

			if cnt == len(docs)-1 || d >= span {
				lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.WiredTiger.Cache.MaxBytesConfigured/acm,
					stat2.WiredTiger.Cache.CurrentlyInCache/acm,
					stat2.WiredTiger.Cache.TrackedDirtyBytes/acm,
					stat2.WiredTiger.Cache.ModifiedPagesEvicted/acm,
					stat2.WiredTiger.Cache.UnmodifiedPagesEvicted/acm,
					stat2.WiredTiger.Cache.PagesReadIntoCache/acm,
					stat2.WiredTiger.Cache.PagesWrittenFromCache/acm))
				acm = 0
				stat1 = stat2
				stat2 = ServerStatusDoc{}
				stat2.Host = stat1.Host
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	return strings.Join(lines, "\n")
}

// printWiredTigerConcurrentTransactionsDetails prints wiredTiger concurrentTransactions stats
func printWiredTigerConcurrentTransactionsDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat := ServerStatusDoc{}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	acm := 0
	lines = append(lines, "\n--- WiredTiger Concurrent Transactions Summary ---")
	lines = append(lines, "+-------------------------+--------------------------------------------+--------------------------------------------+")
	lines = append(lines, "|                         | Read Ticket                                | Write Ticket                               |")
	lines = append(lines, "| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |")
	lines = append(lines, "|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if cnt == 0 {
			stat1 = stat
			stat2.Host = stat1.Host
		} else if stat2.Host == stat.Host {
			d := int(stat.LocalTime.Sub(stat1.LocalTime).Seconds())
			acm++
			stat2.LocalTime = stat.LocalTime
			stat2.WiredTiger.ConcurrentTransactions.Read.Available += stat.WiredTiger.ConcurrentTransactions.Read.Available
			stat2.WiredTiger.ConcurrentTransactions.Read.Out += stat.WiredTiger.ConcurrentTransactions.Read.Out
			stat2.WiredTiger.ConcurrentTransactions.Read.TotalTickets += stat.WiredTiger.ConcurrentTransactions.Read.TotalTickets
			stat2.WiredTiger.ConcurrentTransactions.Write.Available += stat.WiredTiger.ConcurrentTransactions.Write.Available
			stat2.WiredTiger.ConcurrentTransactions.Write.Out += stat.WiredTiger.ConcurrentTransactions.Write.Out
			stat2.WiredTiger.ConcurrentTransactions.Write.TotalTickets += stat.WiredTiger.ConcurrentTransactions.Write.TotalTickets
			if cnt == len(docs)-1 || d >= span {
				lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.WiredTiger.ConcurrentTransactions.Read.Available/acm,
					stat2.WiredTiger.ConcurrentTransactions.Read.Out/acm,
					stat2.WiredTiger.ConcurrentTransactions.Read.TotalTickets/acm,
					stat2.WiredTiger.ConcurrentTransactions.Write.Available/acm,
					stat2.WiredTiger.ConcurrentTransactions.Write.Out/acm,
					stat2.WiredTiger.ConcurrentTransactions.Write.TotalTickets/acm))
				acm = 0
				stat1 = stat2
				stat2 = ServerStatusDoc{}
				stat2.Host = stat1.Host
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	return strings.Join(lines, "\n")
}