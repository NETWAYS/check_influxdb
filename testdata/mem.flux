from(bucket:"monitoring")
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "mem")
        |> aggregateWindow(every: 1h, fn: mean)