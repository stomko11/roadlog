import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../api'

interface Fillup { id: number; date: string; odometer: number; fuelAmount: number; pricePerUnit: number; totalCost: number; station: string; fullTank: boolean }
interface Stats { totalFillups: number; totalSpent: number; avgConsumption: number; totalDistance: number }

export function VehicleDetail() {
  const { id } = useParams()
  const [fillups, setFillups] = useState<Fillup[]>([])
  const [stats, setStats] = useState<Stats | null>(null)

  useEffect(() => {
    api<Fillup[]>(`/vehicles/${id}/fillups`).then(setFillups)
    api<Stats>(`/vehicles/${id}/stats`).then(setStats)
  }, [id])

  return (
    <div className="container">
      <div className="header">
        <Link to="/" className="back">← Back</Link>
        <Link to={`/vehicles/${id}/fillup`} className="btn btn-primary">+ Fill-up</Link>
      </div>

      {stats && (
        <div className="stats">
          <div className="stat-card"><div className="value">{stats.totalFillups}</div><div className="label">Fill-ups</div></div>
          <div className="stat-card"><div className="value">€{stats.totalSpent.toFixed(0)}</div><div className="label">Total Spent</div></div>
          <div className="stat-card"><div className="value">{stats.avgConsumption.toFixed(1)}</div><div className="label">L/100km</div></div>
          <div className="stat-card"><div className="value">{stats.totalDistance.toFixed(0)}</div><div className="label">km driven</div></div>
        </div>
      )}

      <h2>Fill-up History</h2>
      {fillups.length === 0 && <p className="empty">No fill-ups recorded yet.</p>}
      {fillups.map(f => (
        <div key={f.id} className="card" style={{cursor: 'default'}}>
          <div style={{display: 'flex', justifyContent: 'space-between'}}>
            <strong>{new Date(f.date).toLocaleDateString()}</strong>
            <span style={{color: 'var(--accent)'}}>€{f.totalCost.toFixed(2)}</span>
          </div>
          <div style={{fontSize: '0.85rem', color: 'var(--muted)'}}>
            {f.odometer.toFixed(0)} km · {f.fuelAmount.toFixed(1)}L @ €{f.pricePerUnit.toFixed(3)} · {f.station || '—'}
          </div>
        </div>
      ))}
    </div>
  )
}
