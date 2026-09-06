import { useEffect, useState } from 'react'

const emptyForm = { name: '', email: '', phone: '', password: '' }

async function api(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.error || '请求失败，请稍后再试')
  return body
}

function App() {
  const [view, setView] = useState('login')
  const [form, setForm] = useState(emptyForm)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [user, setUser] = useState(null)
  const [cards, setCards] = useState([])
  const [notifications, setNotifications] = useState([])
  const [rooms, setRooms] = useState([])
  const [equipment, setEquipment] = useState([])
  const [loading, setLoading] = useState(false)
  const params = new URLSearchParams(window.location.search)

  useEffect(() => {
    if (window.location.pathname === '/verify-email') {
      const token = params.get('token')
      if (!token) { setError('验证链接缺少 token'); return }
      api(`/api/auth/verify?token=${encodeURIComponent(token)}`)
        .then((data) => { setMessage(data.message); setView('login') })
        .catch((err) => setError(err.message))
    } else if (window.location.pathname === '/reset-password') {
      setView('reset')
      setForm((old) => ({ ...old, token: params.get('token') || '' }))
    }
    api('/api/auth/me').then((data) => setUser(data.user)).catch(() => {})
  }, [])

  useEffect(() => {
    if (user?.role === 'customer') api('/api/customer/membership/cards').then((data) => setCards(data.cards || [])).catch(() => setCards([]))
    else setCards([])
  }, [user])

  useEffect(() => {
    if (user) api('/api/notifications').then((data) => setNotifications(data.notifications || [])).catch(() => setNotifications([]))
    else setNotifications([])
  }, [user])

  useEffect(() => {
    Promise.all([api('/api/rooms'), api('/api/equipment')]).then(([roomData, equipmentData]) => {
      setRooms(roomData.rooms || []); setEquipment(equipmentData.equipment || [])
    }).catch(() => {})
  }, [])

  function updateField(event) {
    setForm((old) => ({ ...old, [event.target.name]: event.target.value }))
    setError('')
  }

  async function submit(event) {
    event.preventDefault()
    setLoading(true); setError(''); setMessage('')
    try {
      if (view === 'login') {
        const data = await api('/api/auth/login', { method: 'POST', body: JSON.stringify({ email: form.email, password: form.password }) })
        setUser(data.user); setMessage(`欢迎回来，${data.user.name}`)
      } else if (view === 'register') {
        const data = await api('/api/auth/register', { method: 'POST', body: JSON.stringify(form) })
        setMessage(data.message); setView('login'); setForm(emptyForm)
      } else if (view === 'forgot') {
        const data = await api('/api/auth/password-reset/request', { method: 'POST', body: JSON.stringify({ email: form.email }) })
        setMessage(data.message)
      } else {
        const data = await api('/api/auth/password-reset/confirm', { method: 'POST', body: JSON.stringify({ token: form.token, password: form.password }) })
        setMessage(data.message); setView('login'); setForm(emptyForm)
      }
    } catch (err) { setError(err.message) } finally { setLoading(false) }
  }

  async function logout() {
    await api('/api/auth/logout', { method: 'POST' }).catch(() => {})
    setUser(null); setCards([]); setNotifications([]); setMessage('已退出登录')
  }

  const titles = { login: '登录 BandRoom', register: '创建顾客账号', forgot: '找回密码', reset: '设置新密码' }
  const isReset = view === 'reset'

  return (
    <main className="shell">
      <div className="ambient ambient-one" /><div className="ambient ambient-two" />
      <section className="brand-panel">
        <p className="eyebrow">BANDROOM · MUSIC SPACE</p>
        <h1>让每一次排练，<span>准时开始。</span></h1>
        <p className="lead">多房间预约、会员卡与排练设备，一处管理你的乐队时光。</p>
        <div className="feature-list"><span>01　灵活预约</span><span>02　设备借用</span><span>03　会员专属</span></div>
        <Catalog rooms={rooms} equipment={equipment} />
      </section>
      <section className="auth-card">
        <div className="card-top"><span className="record-dot" /><span>REHEARSAL ROOM ACCESS</span></div>
        {user ? <LoggedIn user={user} cards={cards} notifications={notifications} setNotifications={setNotifications} rooms={rooms} equipment={equipment} onLogout={logout} /> : <>
          <div className="card-heading"><p className="eyebrow">ACCOUNT</p><h2>{titles[view]}</h2><p>{view === 'register' ? '注册后即可查看空闲房间并提交预约。' : '登录后管理你的排练预约与会员权益。'}</p></div>
          {message && <div className="notice success">{message}</div>}
          {error && <div className="notice error">{error}</div>}
          <form onSubmit={submit}>
            {view === 'register' && <><Field label="联系人姓名" name="name" value={form.name} onChange={updateField} required /><Field label="手机号" name="phone" value={form.phone} onChange={updateField} required /></>}
            {isReset && <Field label="重置令牌" name="token" value={form.token || ''} onChange={updateField} required />}
            {view !== 'reset' && <Field label="邮箱" name="email" type="email" value={form.email} onChange={updateField} required />}
            {view !== 'forgot' && <Field label={isReset ? '新密码' : '密码'} name="password" type="password" value={form.password} onChange={updateField} required />}
            {view === 'register' && <p className="hint">密码至少 8 位；注册后请从演示邮件中打开验证链接。</p>}
            <button className="primary-button" disabled={loading}>{loading ? '处理中…' : view === 'login' ? '登录' : view === 'register' ? '注册并验证邮箱' : view === 'forgot' ? '发送重置邮件' : '确认新密码'}</button>
          </form>
          <AuthLinks view={view} onChange={(next) => { setView(next); setError(''); setMessage(''); setForm(emptyForm) }} />
        </>}
      </section>
    </main>
  )
}

function Field({ label, name, type = 'text', value, onChange, required }) {
  return <label className="field"><span>{label}</span><input name={name} type={type} value={value} onChange={onChange} required={required} autoComplete={type === 'password' ? 'new-password' : name} /></label>
}

function AuthLinks({ view, onChange }) {
  if (view === 'reset') return <p className="switch-line"><button onClick={() => onChange('login')}>返回登录</button></p>
  return <div className="switch-line">{view === 'login' ? <><span>还没有账号？</span><button onClick={() => onChange('register')}>立即注册</button><button onClick={() => onChange('forgot')}>忘记密码</button></> : <button onClick={() => onChange('login')}>返回登录</button>}</div>
}

function LoggedIn({ user, cards, notifications, setNotifications, rooms, equipment, onLogout }) {
  return <div className="logged-in"><div className="avatar">{user.name.slice(0, 1)}</div><h2>你好，{user.name}</h2><p>{user.email}</p><div className="member-badge">{user.role === 'owner' ? '老板账户' : user.emailVerifiedAt ? '邮箱已验证' : '待验证'}</div><Notifications items={notifications} setItems={setNotifications} />{user.role === 'customer' && <><MembershipCards cards={cards} /><BookingCenter user={user} rooms={rooms} equipment={equipment} /></>}<button className="secondary-button">进入预约中心</button><button className="text-button" onClick={onLogout}>退出登录</button></div>
}

function MembershipCards({ cards }) {
  if (!cards.length) return <div className="membership-empty">暂时没有会员卡，请联系老板线下开通。</div>
  return <div className="membership-list">{cards.map((card) => <div className="membership-card" key={card.id}><div><strong>{card.planName}</strong><span>{card.planType === 'monthly' ? '月卡' : '次数卡'}</span></div><b>{card.planType === 'monthly' ? `${card.startsAt.slice(0, 10)} — ${card.endsAt.slice(0, 10)}` : `剩余 ${card.remainingUses ?? 0} 次`}</b><small>{card.status === 'active' ? '当前可用' : card.status === 'scheduled' ? '待生效' : card.status === 'depleted' ? '已用完' : card.status}</small></div>)}</div>
}

function Notifications({ items, setItems }) {
  async function markRead(id) { await api(`/api/notifications/${id}/read`, { method: 'PATCH' }).catch(() => {}); setItems((old) => old.map((item) => item.id === id ? { ...item, readAt: new Date().toISOString() } : item)) }
  if (!items.length) return <div className="notification-empty">暂无站内通知</div>
  return <div className="notifications"><div className="booking-heading"><span>通知</span><small>{items.filter((item) => !item.readAt).length} 条未读</small></div>{items.slice(0, 3).map((item) => <button type="button" className={item.readAt ? 'notification read' : 'notification'} key={item.id} onClick={() => markRead(item.id)}><b>{item.title}</b><span>{item.body}</span></button>)}</div>
}

function Catalog({ rooms, equipment }) {
  return <div className="catalog"><div className="catalog-heading"><span>空间与设备</span><small>开放浏览 · 登录后预约</small></div><div className="catalog-grid">{rooms.slice(0, 3).map((room) => <div className="catalog-item" key={room.id}><strong>{room.name}</strong><span>{room.status === 'available' ? '可预约' : '暂不可用'} · {room.capacity ? `${room.capacity} 人` : '标准房'}</span><small>{room.fixedEquipment?.map((item) => item.name).join(' · ') || '基础排练设备'}</small></div>)}{equipment.slice(0, 3).map((item) => <div className="catalog-item equipment-item" key={`equipment-${item.id}`}><strong>{item.name}</strong><span>{item.status === 'available' ? `库存 ${item.quantity}` : '维护中'}</span><small>公共借用设备</small></div>)}</div></div>
}

function BookingCenter({ user, rooms, equipment }) {
  const [date, setDate] = useState(tomorrow())
  const [roomID, setRoomID] = useState('')
  const [slots, setSlots] = useState([])
  const [selectedSlot, setSelectedSlot] = useState(null)
  const [details, setDetails] = useState({ bandName: '', phone: user.phone || '', notes: '' })
  const [quantities, setQuantities] = useState({})
  const [bookings, setBookings] = useState([])
  const [bookingMessage, setBookingMessage] = useState('')
  const [bookingError, setBookingError] = useState('')
  const availableRooms = rooms.filter((room) => room.status === 'available')

  useEffect(() => { if (!roomID && availableRooms[0]) setRoomID(String(availableRooms[0].id)) }, [rooms])
  useEffect(() => {
    if (!roomID || !date) return
    setSelectedSlot(null); setBookingError('')
    api(`/api/availability?roomId=${roomID}&date=${date}`).then((data) => setSlots(data.slots || [])).catch((err) => setBookingError(err.message))
  }, [roomID, date])
  useEffect(() => { api('/api/customer/bookings').then((data) => setBookings(data.bookings || [])).catch(() => {}) }, [])

  async function submitBooking(event) {
    event.preventDefault(); setBookingMessage(''); setBookingError('')
    if (!selectedSlot) { setBookingError('请先选择一个可用时段'); return }
    try {
      const selectedEquipment = Object.entries(quantities).filter(([, quantity]) => Number(quantity) > 0).map(([equipmentId, quantity]) => ({ equipmentId: Number(equipmentId), quantity: Number(quantity) }))
      const data = await api('/api/customer/bookings', { method: 'POST', body: JSON.stringify({ roomId: Number(roomID), ...details, startsAt: selectedSlot.startsAt, endsAt: selectedSlot.endsAt, equipment: selectedEquipment }) })
      setBookingMessage(`预约成功：${data.booking.startsAt.slice(0, 16).replace('T', ' ')} 开始`); setBookings((old) => [data.booking, ...old]); setSelectedSlot(null)
    } catch (err) { setBookingError(err.message) }
  }

  async function cancelBooking(bookingID) {
    try {
      await api(`/api/customer/bookings/${bookingID}/cancel`, { method: 'POST', body: JSON.stringify({ reason: '顾客主动取消' }) })
      setBookings((old) => old.map((booking) => booking.id === bookingID ? { ...booking, status: 'cancelled' } : booking))
      setBookingMessage('预约已取消，房间和次数已释放。')
    } catch (err) { setBookingError(err.message) }
  }

  return <div className="booking-center"><div className="booking-heading"><span>预约排练房</span><small>选择日期 · 房间 · 时段</small></div>{availableRooms.length === 0 ? <div className="membership-empty">当前没有可预约的房间。</div> : <form onSubmit={submitBooking}><div className="booking-fields"><label className="field"><span>预约日期</span><input type="date" value={date} min={today()} onChange={(event) => setDate(event.target.value)} /></label><label className="field"><span>排练房</span><select value={roomID} onChange={(event) => setRoomID(event.target.value)}>{availableRooms.map((room) => <option value={room.id} key={room.id}>{room.name}</option>)}</select></label></div><div className="slot-picker"><span className="field-label">可用时段（含 30 分钟清场）</span><div className="slot-grid">{slots.map((slot) => <button type="button" className={selectedSlot?.startsAt === slot.startsAt && selectedSlot?.endsAt === slot.endsAt ? 'slot selected' : 'slot'} key={`${slot.startsAt}-${slot.endsAt}`} onClick={() => setSelectedSlot(slot)}>{slot.startsAt.slice(11, 16)}–{slot.endsAt.slice(11, 16)}<small>{slot.durationMinutes / 60} 小时</small></button>)}{!slots.length && <small className="muted">当天没有符合条件的时段</small>}</div></div><div className="booking-fields"><Field label="乐队名称" name="bandName" value={details.bandName} onChange={(event) => setDetails({ ...details, bandName: event.target.value })} required /><Field label="到店手机号" name="phone" value={details.phone} onChange={(event) => setDetails({ ...details, phone: event.target.value })} required /></div><label className="field"><span>备注</span><input value={details.notes} onChange={(event) => setDetails({ ...details, notes: event.target.value })} placeholder="可选" /></label><div className="equipment-picker"><span className="field-label">额外借用设备（不额外收费）</span>{equipment.filter((item) => item.status === 'available' && item.quantity > 0).map((item) => <label key={item.id}><span>{item.name} · 库存 {item.quantity}</span><select value={quantities[item.id] || 0} onChange={(event) => setQuantities({ ...quantities, [item.id]: event.target.value })}>{Array.from({ length: Math.min(item.quantity, 5) + 1 }, (_, index) => <option value={index} key={index}>{index === 0 ? '不借用' : `${index} 件`}</option>)}</select></label>)}</div>{bookingMessage && <div className="notice success">{bookingMessage}</div>}{bookingError && <div className="notice error">{bookingError}</div>}<button className="primary-button" type="submit">立即预约</button></form>}<div className="booking-history"><span className="field-label">我的预约</span>{bookings.slice(0, 3).map((booking) => <div className="history-item" key={booking.id}><b>{booking.startsAt.slice(0, 16).replace('T', ' ')}</b><small>{booking.bandName} · {booking.status === 'confirmed' ? '已确认' : booking.status}</small>{booking.status === 'confirmed' && <button type="button" className="cancel-button" onClick={() => cancelBooking(booking.id)}>取消</button>}</div>)}</div></div>
}

function today() { return new Date().toISOString().slice(0, 10) }
function tomorrow() { const date = new Date(); date.setDate(date.getDate() + 1); return date.toISOString().slice(0, 10) }

export default App
