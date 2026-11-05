/**
 * Servidor API de Desenvolvimento Local para Teste Manual Imediato
 * Porta 8080 - Sem dependências externas (apenas Node.js nativo)
 */
const http = require("http");
const https = require("https");
const { randomUUID } = require("crypto");

const PORT = 8080;

// Estado em memória com dados semente para teste imediato
let producers = [
  // PRODUTORES DO CEARÁ (Diversos polos agrícolas e relevos)
  {
    id: "prod-ce-1",
    name: "Dona Francisca Chagas",
    whatsapp_phone: "5588998112233",
    city: "Tianguá",
    state: "CE",
    created_at: new Date(Date.now() - 86400000 * 6).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prod-ce-2",
    name: "Antônio Ferreira (Agrícola Vale Verde)",
    whatsapp_phone: "5588997223344",
    city: "Limoeiro do Norte",
    state: "CE",
    created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prod-ce-3",
    name: "Seu Raimundo Nonato do Cariri",
    whatsapp_phone: "5588996334455",
    city: "Crato",
    state: "CE",
    created_at: new Date(Date.now() - 86400000 * 4).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prod-ce-4",
    name: "Zé Carlos do Caju",
    whatsapp_phone: "5585995445566",
    city: "Cascavel",
    state: "CE",
    created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prod-ce-5",
    name: "Maria de Fátima Bezerra",
    whatsapp_phone: "5588994556677",
    city: "Quixadá",
    state: "CE",
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date().toISOString(),
  },
  // PRODUTORES SUDESTE (SP / MG)
  {
    id: "prod-1",
    name: "Seu Zé da Silva",
    whatsapp_phone: "5519987654321",
    city: "Campinas",
    state: "SP",
    created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prod-2",
    name: "Dona Ana Oliveira",
    whatsapp_phone: "5519976543210",
    city: "Piracicaba",
    state: "SP",
    created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prod-3",
    name: "João da Mata",
    whatsapp_phone: "5535991234567",
    city: "Guaxupé",
    state: "MG",
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

let properties = [
  // PROPRIEDADES NO CEARÁ
  // 1. Serra da Ibiapaba (Planalto 750m - Clima ameno e relevo montanhoso)
  {
    id: "prop-ce-1",
    producer_id: "prod-ce-1",
    latitude: -3.7317,
    longitude: -40.9922,
    crop: "tomate",
    soil_type: "argiloso",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 6).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-ce-2",
    producer_id: "prod-ce-1",
    latitude: -3.8541,
    longitude: -40.9234,
    crop: "cafe",
    soil_type: "medio",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 6).toISOString(),
    updated_at: new Date().toISOString(),
  },
  // 2. Vale do Jaguaribe (Planície fluvial irrigada e baixio)
  {
    id: "prop-ce-3",
    producer_id: "prod-ce-2",
    latitude: -5.1458,
    longitude: -38.0983,
    crop: "fruticultura",
    soil_type: "aluvial",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-ce-4",
    producer_id: "prod-ce-2",
    latitude: -5.0744,
    longitude: -37.9881,
    crop: "soja",
    soil_type: "aluvial",
    crop_stage: "germination",
    created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
    updated_at: new Date().toISOString(),
  },
  // 3. Região do Cariri (Pé da Chapada do Araripe)
  {
    id: "prop-ce-5",
    producer_id: "prod-ce-3",
    latitude: -7.2339,
    longitude: -39.4103,
    crop: "milho",
    soil_type: "medio",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 4).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-ce-6",
    producer_id: "prod-ce-3",
    latitude: -7.3128,
    longitude: -39.3047,
    crop: "mandioca",
    soil_type: "medio",
    crop_stage: "germination",
    created_at: new Date(Date.now() - 86400000 * 4).toISOString(),
    updated_at: new Date().toISOString(),
  },
  // 4. Litoral Leste / Tabuleiros Costeiros
  {
    id: "prop-ce-7",
    producer_id: "prod-ce-4",
    latitude: -4.1311,
    longitude: -38.2439,
    crop: "caju",
    soil_type: "arenoso",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
    updated_at: new Date().toISOString(),
  },
  // 5. Sertão Central (Depressão sertaneja com relevo de inselbergs e solo raso)
  {
    id: "prop-ce-8",
    producer_id: "prod-ce-5",
    latitude: -4.9714,
    longitude: -39.0156,
    crop: "feijao",
    soil_type: "raso",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-ce-9",
    producer_id: "prod-ce-5",
    latitude: -5.1983,
    longitude: -39.2933,
    crop: "algodao",
    soil_type: "raso",
    crop_stage: "germination",
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date().toISOString(),
  },

  // PROPRIEDADES SUDESTE
  {
    id: "prop-1",
    producer_id: "prod-1",
    latitude: -22.9068,
    longitude: -47.0616,
    crop: "milho",
    soil_type: "arenoso",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-2",
    producer_id: "prod-1",
    latitude: -22.8200,
    longitude: -47.1200,
    crop: "tomate",
    soil_type: "medio",
    crop_stage: "germination",
    created_at: new Date(Date.now() - 86400000 * 4).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-3",
    producer_id: "prod-2",
    latitude: -22.7253,
    longitude: -47.6492,
    crop: "milho",
    soil_type: "argiloso",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "prop-4",
    producer_id: "prod-3",
    latitude: -21.3050,
    longitude: -46.7125,
    crop: "cafe",
    soil_type: "argiloso",
    crop_stage: "flowering",
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

let alerts = [
  // ALERTAS ATIVOS NO CEARÁ
  {
    id: "alt-ce-1",
    property_id: "prop-ce-8",
    level: "critical",
    alert_type: "severe_water_stress",
    period_start: new Date(Date.now() + 86400000).toISOString(),
    period_end: new Date(Date.now() + 86400000 * 3).toISOString(),
    triggered_at: new Date(Date.now() - 3600000 * 2).toISOString(),
    resolved_at: null,
  },
  {
    id: "alt-ce-2",
    property_id: "prop-ce-7",
    level: "medium",
    alert_type: "heat_wave",
    period_start: new Date(Date.now() + 86400000 * 2).toISOString(),
    period_end: new Date(Date.now() + 86400000 * 4).toISOString(),
    triggered_at: new Date(Date.now() - 3600000 * 5).toISOString(),
    resolved_at: null,
  },
  {
    id: "alt-ce-3",
    property_id: "prop-ce-1",
    level: "medium",
    alert_type: "heavy_rain",
    period_start: new Date(Date.now() + 86400000).toISOString(),
    period_end: new Date(Date.now() + 86400000 * 2).toISOString(),
    triggered_at: new Date(Date.now() - 3600000 * 6).toISOString(),
    resolved_at: null,
  },
  // ALERTAS SUDESTE
  {
    id: "alt-1",
    property_id: "prop-1",
    level: "critical",
    alert_type: "severe_water_stress",
    period_start: new Date(Date.now() + 86400000).toISOString(),
    period_end: new Date(Date.now() + 86400000 * 3).toISOString(),
    triggered_at: new Date(Date.now() - 3600000 * 4).toISOString(),
    resolved_at: null,
  },
  {
    id: "alt-2",
    property_id: "prop-3",
    level: "medium",
    alert_type: "heavy_rain",
    period_start: new Date(Date.now() + 86400000 * 2).toISOString(),
    period_end: new Date(Date.now() + 86400000 * 3).toISOString(),
    triggered_at: new Date(Date.now() - 3600000 * 8).toISOString(),
    resolved_at: null,
  },
];

let notifications = [
  // NOTIFICAÇÕES CEARÁ
  {
    id: "notif-ce-1",
    alert_id: "alt-ce-1",
    producer_id: "prod-ce-5",
    property_id: "prop-ce-8",
    phone: "5588994556677",
    message: "Alerta Crítico: Estresse hídrico severo previsto em Quixadá para os próximos 3 dias. Solo raso do sertão atinge capacidade mínima. Recomendação: Priorize a irrigação pontual de emergência na floração do feijão-caupi.",
    status: "read",
    sent_at: new Date(Date.now() - 3600000 * 2).toISOString(),
    delivered_at: new Date(Date.now() - 3600000 * 2 + 4000).toISOString(),
    read_at: new Date(Date.now() - 3600000 * 1).toISOString(),
  },
  {
    id: "notif-ce-2",
    alert_id: "alt-ce-2",
    producer_id: "prod-ce-4",
    property_id: "prop-ce-7",
    phone: "5585995445566",
    message: "Alerta Médio: Onda de calor com picos > 36°C no litoral de Cascavel. Solo arenoso com rápida perda por evapotranspiração. Recomendação: Mantenha cobertura vegetal morta sobre os canteiros do cajueiro.",
    status: "delivered",
    sent_at: new Date(Date.now() - 3600000 * 5).toISOString(),
    delivered_at: new Date(Date.now() - 3600000 * 5 + 3000).toISOString(),
    read_at: null,
  },
  {
    id: "notif-ce-3",
    alert_id: "alt-ce-3",
    producer_id: "prod-ce-1",
    property_id: "prop-ce-1",
    phone: "5588998112233",
    message: "Alerta Médio: Chuva forte prevista para a Serra da Ibiapaba (> 40mm). Risco de encharcamento em solo argiloso. Recomendação: Desobstrua canais de drenagem e suspenda a fertirrigação.",
    status: "sent",
    sent_at: new Date(Date.now() - 3600000 * 6).toISOString(),
    delivered_at: null,
    read_at: null,
  },
  // NOTIFICAÇÕES SUDESTE
  {
    id: "notif-1",
    alert_id: "alt-1",
    producer_id: "prod-1",
    property_id: "prop-1",
    phone: "5519987654321",
    message: "Risco de estresse hídrico severo detectado para os próximos 3 dias.\nRecomendação: Aumente o turno de irrigação para o período noturno.",
    status: "read",
    sent_at: new Date(Date.now() - 3600000 * 4).toISOString(),
    delivered_at: new Date(Date.now() - 3600000 * 4 + 5000).toISOString(),
    read_at: new Date(Date.now() - 3600000 * 3).toISOString(),
  },
  {
    id: "notif-2",
    alert_id: "alt-2",
    producer_id: "prod-2",
    property_id: "prop-3",
    phone: "5519976543210",
    message: "Chuva forte prevista (soma acumulada > 35mm).\nRecomendação: Verifique os drenos e proteja as áreas baixas.",
    status: "delivered",
    sent_at: new Date(Date.now() - 3600000 * 8).toISOString(),
    delivered_at: new Date(Date.now() - 3600000 * 8 + 3000).toISOString(),
    read_at: null,
  },
];

function sendJson(res, statusCode, data) {
  res.writeHead(statusCode, {
    "Content-Type": "application/json",
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type, Authorization",
  });
  res.end(JSON.stringify(data));
}

function parseBody(req) {
  return new Promise((resolve, reject) => {
    let body = "";
    req.on("data", (chunk) => (body += chunk));
    req.on("end", () => {
      try {
        resolve(body ? JSON.parse(body) : {});
      } catch (err) {
        reject(err);
      }
    });
    req.on("error", reject);
  });
}

function fetchOpenMeteo(lat, lon) {
  return new Promise((resolve) => {
    const url = `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&hourly=temperature_2m,relative_humidity_2m,precipitation,wind_speed_10m,cloud_cover,weather_code,apparent_temperature,precipitation_probability`;
    https
      .get(url, (res) => {
        let raw = "";
        res.on("data", (c) => (raw += c));
        res.on("end", () => {
          try {
            const data = JSON.parse(raw);
            if (!data.hourly) throw new Error("no hourly");
            const readings = data.hourly.time.map((t, i) => ({
              time: t,
              temp_avg_c: data.hourly.temperature_2m[i],
              humidity_pct: data.hourly.relative_humidity_2m[i],
              precipitation_mm: data.hourly.precipitation[i],
              cloud_cover_pct: data.hourly.cloud_cover ? data.hourly.cloud_cover[i] : 25,
              wind_speed_kmh: data.hourly.wind_speed_10m ? data.hourly.wind_speed_10m[i] : 12,
              weather_code: data.hourly.weather_code ? data.hourly.weather_code[i] : 1,
              apparent_temp_c: data.hourly.apparent_temperature ? data.hourly.apparent_temperature[i] : data.hourly.temperature_2m[i],
              rain_prob_pct: data.hourly.precipitation_probability ? data.hourly.precipitation_probability[i] : (data.hourly.precipitation[i] > 0 ? 80 : 10),
            }));
            resolve(readings);
          } catch {
            resolve(mockForecast());
          }
        });
      })
      .on("error", () => resolve(mockForecast()));
  });
}

function mockForecast() {
  const readings = [];
  const now = new Date();
  for (let d = 0; d < 7; d++) {
    for (let h = 0; h < 24; h += 4) {
      const t = new Date(now.getTime() + (d * 24 + h) * 3600000);
      const isRainy = d === 2;
      const cloudPct = isRainy ? 85 : Math.round(20 + Math.sin(d + h) * 30);
      readings.push({
        time: t.toISOString().slice(0, 16),
        temp_avg_c: 24 + Math.sin(h / 4) * 8,
        humidity_pct: 60 - Math.sin(h / 4) * 20,
        precipitation_mm: isRainy ? 14.5 : 0,
        cloud_cover_pct: cloudPct,
        wind_speed_kmh: 12 + Math.round(Math.sin(h) * 6),
        weather_code: isRainy ? 63 : (cloudPct > 60 ? 3 : (cloudPct > 30 ? 2 : 0)),
        apparent_temp_c: 25 + Math.sin(h / 4) * 9,
        rain_prob_pct: isRainy ? 90 : 5,
      });
    }
  }
  return readings;
}

const server = http.createServer(async (req, res) => {
  const parsedUrl = new URL(req.url, `http://${req.headers.host}`);
  const pathname = parsedUrl.pathname;
  const method = req.method;

  if (method === "OPTIONS") {
    res.writeHead(204, {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type, Authorization",
    });
    return res.end();
  }

  try {
    // HEALTH CHECK
    if (pathname === "/health" && method === "GET") {
      return sendJson(res, 200, { status: "ok" });
    }

    // SCAN SOB DEMANDA
    if (pathname === "/scan/run" && method === "POST") {
      return sendJson(res, 200, { status: "scan_completed", scanned: properties.length });
    }

    // NOTIFICATIONS
    if (pathname === "/notifications" && method === "GET") {
      return sendJson(res, 200, notifications);
    }

    // WEBHOOK EVOLUTION
    if (pathname === "/webhooks/evolution" && method === "POST") {
      const body = await parseBody(req);
      const phone = (body?.data?.key?.remoteJid || "").split("@")[0];
      if (phone) {
        const notif = notifications.find((n) => n.phone === phone);
        if (notif) {
          const st = (body?.data?.status || "").toUpperCase();
          if (st.includes("READ")) notif.status = "read";
          else if (st.includes("DELIVER")) notif.status = "delivered";
        }
      }
      return sendJson(res, 200, { status: "updated" });
    }

    // ALERTS
    if (pathname === "/alerts" && method === "GET") {
      return sendJson(res, 200, alerts);
    }
    const resolveMatch = pathname.match(/^\/alerts\/([^/]+)\/resolve$/);
    if (resolveMatch && method === "POST") {
      const id = resolveMatch[1];
      const alt = alerts.find((a) => a.id === id);
      if (!alt) return sendJson(res, 404, { error: "alert not found" });
      alt.resolved_at = new Date().toISOString();
      return sendJson(res, 200, { status: "resolved", id });
    }

    // PRODUCERS
    if (pathname === "/producers" && method === "GET") {
      return sendJson(res, 200, producers);
    }
    if (pathname === "/producers" && method === "POST") {
      const body = await parseBody(req);
      const newProd = {
        id: randomUUID(),
        name: body.name || "Novo Produtor",
        whatsapp_phone: body.whatsapp_phone || "",
        city: body.city || "",
        state: body.state || "",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      producers.unshift(newProd);
      return sendJson(res, 201, newProd);
    }
    const producerMatch = pathname.match(/^\/producers\/([^/]+)$/);
    if (producerMatch) {
      const id = producerMatch[1];
      const prod = producers.find((p) => p.id === id);
      if (method === "GET") {
        if (!prod) return sendJson(res, 404, { error: "producer not found" });
        return sendJson(res, 200, prod);
      }
      if (method === "PUT") {
        if (!prod) return sendJson(res, 404, { error: "producer not found" });
        const body = await parseBody(req);
        Object.assign(prod, body, { updated_at: new Date().toISOString() });
        return sendJson(res, 200, prod);
      }
      if (method === "DELETE") {
        producers = producers.filter((p) => p.id !== id);
        properties = properties.filter((p) => p.producer_id !== id);
        return sendJson(res, 200, { status: "deleted", id });
      }
    }
    const producerPropsMatch = pathname.match(/^\/producers\/([^/]+)\/properties$/);
    if (producerPropsMatch && method === "GET") {
      const id = producerPropsMatch[1];
      return sendJson(
        res,
        200,
        properties.filter((p) => p.producer_id === id)
      );
    }
    const producerNotifsMatch = pathname.match(/^\/producers\/([^/]+)\/notifications$/);
    if (producerNotifsMatch && method === "GET") {
      const id = producerNotifsMatch[1];
      return sendJson(
        res,
        200,
        notifications.filter((n) => n.producer_id === id)
      );
    }

    // PROPERTIES
    if (pathname === "/properties" && method === "GET") {
      const producerId = parsedUrl.searchParams.get("producer_id");
      if (producerId) {
        return sendJson(
          res,
          200,
          properties.filter((p) => p.producer_id === producerId)
        );
      }
      return sendJson(res, 200, properties);
    }
    if (pathname === "/properties" && method === "POST") {
      const body = await parseBody(req);
      const newProp = {
        id: randomUUID(),
        producer_id: body.producer_id,
        latitude: Number(body.latitude),
        longitude: Number(body.longitude),
        crop: body.crop || "milho",
        soil_type: body.soil_type || "medio",
        crop_stage: body.crop_stage || "flowering",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      properties.unshift(newProp);
      return sendJson(res, 201, newProp);
    }
    const propMatch = pathname.match(/^\/properties\/([^/]+)$/);
    if (propMatch) {
      const id = propMatch[1];
      const prop = properties.find((p) => p.id === id);
      if (method === "GET") {
        if (!prop) return sendJson(res, 404, { error: "property not found" });
        return sendJson(res, 200, prop);
      }
      if (method === "PUT") {
        if (!prop) return sendJson(res, 404, { error: "property not found" });
        const body = await parseBody(req);
        Object.assign(prop, body, {
          latitude: Number(body.latitude ?? prop.latitude),
          longitude: Number(body.longitude ?? prop.longitude),
          updated_at: new Date().toISOString(),
        });
        return sendJson(res, 200, prop);
      }
      if (method === "DELETE") {
        properties = properties.filter((p) => p.id !== id);
        alerts = alerts.filter((a) => a.property_id !== id);
        return sendJson(res, 200, { status: "deleted", id });
      }
    }
    const propAlertsMatch = pathname.match(/^\/properties\/([^/]+)\/alerts$/);
    if (propAlertsMatch && method === "GET") {
      const id = propAlertsMatch[1];
      return sendJson(
        res,
        200,
        alerts.filter((a) => a.property_id === id)
      );
    }
    const propForecastMatch = pathname.match(/^\/properties\/([^/]+)\/forecast$/);
    if (propForecastMatch && method === "GET") {
      const id = propForecastMatch[1];
      const prop = properties.find((p) => p.id === id);
      if (!prop) return sendJson(res, 404, { error: "property not found" });
      const forecast = await fetchOpenMeteo(prop.latitude, prop.longitude);
      return sendJson(res, 200, forecast);
    }

    sendJson(res, 404, { error: "endpoint not found" });
  } catch (err) {
    sendJson(res, 500, { error: err.message });
  }
});

server.listen(PORT, () => {
  console.log(`[AgroResiliente Dev API] rodando na porta ${PORT}`);
  console.log(`- GET  /producers`);
  console.log(`- GET  /properties`);
  console.log(`- GET  /alerts`);
  console.log(`- GET  /notifications`);
});

