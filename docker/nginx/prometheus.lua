  local prometheus = require "resty.prometheus"
  local metric_connections, metric_connect_time

  local function init()
      prometheus = prometheus.init("prometheus_metrics")
      metric_connections = prometheus:counter("nginx_stream_connections_total", "Total connections", {"upstream"})
      metric_connect_time = prometheus:histogram("nginx_stream_upstream_connect_seconds", "Upstream connect time",
  {"upstream"})
  end

  local function log_connect_time()
      local connect_time = tonumber(ngx.var.upstream_connect_time)
      local upstream = ngx.var.upstream_addr or "unknown"

      if connect_time then
          metric_connect_time:observe(connect_time / 1000, {upstream})  -- Convert ms to seconds
          metric_connections:inc(1, {upstream})
      end
  end

  return {
      init = init,
      log_connect_time = log_connect_time
  }