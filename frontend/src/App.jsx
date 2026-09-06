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
    setUser(null); setCards([]); setMessage('已退出登录')
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
      </section>
      <section className="auth-card">
        <div className="card-top"><span className="record-dot" /><span>REHEARSAL ROOM ACCESS</span></div>
        {user ? <LoggedIn user={user} cards={cards} onLogout={logout} /> : <>
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

function LoggedIn({ user, cards, onLogout }) {
  return <div className="logged-in"><div className="avatar">{user.name.slice(0, 1)}</div><h2>你好，{user.name}</h2><p>{user.email}</p><div className="member-badge">{user.role === 'owner' ? '老板账户' : user.emailVerifiedAt ? '邮箱已验证' : '待验证'}</div>{user.role === 'customer' && <MembershipCards cards={cards} />}<button className="secondary-button">进入预约中心</button><button className="text-button" onClick={onLogout}>退出登录</button></div>
}

function MembershipCards({ cards }) {
  if (!cards.length) return <div className="membership-empty">暂时没有会员卡，请联系老板线下开通。</div>
  return <div className="membership-list">{cards.map((card) => <div className="membership-card" key={card.id}><div><strong>{card.planName}</strong><span>{card.planType === 'monthly' ? '月卡' : '次数卡'}</span></div><b>{card.planType === 'monthly' ? `${card.startsAt.slice(0, 10)} — ${card.endsAt.slice(0, 10)}` : `剩余 ${card.remainingUses ?? 0} 次`}</b><small>{card.status === 'active' ? '当前可用' : card.status === 'scheduled' ? '待生效' : card.status === 'depleted' ? '已用完' : card.status}</small></div>)}</div>
}

export default App
