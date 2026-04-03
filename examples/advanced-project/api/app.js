const express = require('express');
const { Pool } = require('pg');
const { createClient } = require('redis');

const app = express();
const port = process.env.PORT || 3000;

// PostgreSQL Client
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
});

// Redis Client
const redisClient = createClient({
  url: process.env.REDIS_URL,
});

redisClient.on('error', (err) => console.log('Redis Client Error', err));

app.get('/', async (req, res) => {
  try {
    // Check Postgres
    const dbRes = await pool.query('SELECT NOW() as now');
    
    // Check Redis
    if (!redisClient.isOpen) await redisClient.connect();
    await redisClient.set('last_access', new Date().toISOString());
    const lastAccess = await redisClient.get('last_access');

    res.json({
      status: '🚀 Derrick Demo API is running!',
      database: {
        status: 'Connected',
        time: dbRes.rows[0].now
      },
      cache: {
        status: 'Connected',
        last_access: lastAccess
      },
      environment: process.env.NODE_ENV || 'development'
    });
  } catch (error) {
    console.error(error);
    res.status(500).json({ error: error.message });
  }
});

app.listen(port, () => {
  console.log(`📡 API listening at http://localhost:${port}`);
});
