'use client';

import { AppShell } from '@/components/AppShell';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { API_BASE_URL } from '@/lib/api';
import { getToken, useAuthGuard } from '@/lib/auth';
import { Activity, Bell, CircleAlert, Loader2, Mail, Megaphone, Plus, Send, Users } from 'lucide-react';
import { FormEvent, useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

type Overview = { users: number; announcements: number; emails: number };
type Analytics = { totals: { active_cvs: number; new_users_30d: number; new_cvs_30d: number; active_users_7d: number } };
type User = { id: string; full_name: string; email?: string; plan: string; role: string };
type Announcement = { id: string; title: string; subject: string; body: string; in_app: boolean; email: boolean; status: string; created_at: string };
type EmailLog = { id: number; to: string; kind: string; status: string; created_at: string };
type Tab = 'overview' | 'announcements' | 'users' | 'delivery';

async function adminRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}/admin${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${getToken()}`, ...init?.headers },
  });
  const payload = await response.json();
  if (!response.ok || !payload.success) throw new Error(payload.error?.message || 'The admin request failed.');
  return payload.data as T;
}

function dateLabel(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

export default function AdminPage() {
  useAuthGuard();
  const [tab, setTab] = useState<Tab>('overview');
  const [overview, setOverview] = useState<Overview | null>(null);
  const [analytics, setAnalytics] = useState<Analytics | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [logs, setLogs] = useState<EmailLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [query, setQuery] = useState('');
  const [title, setTitle] = useState('');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [inApp, setInApp] = useState(true);
  const [email, setEmail] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [stats, usage, people, notices, mail] = await Promise.all([
        adminRequest<Overview>('/overview'),
        adminRequest<Analytics>('/analytics'),
        adminRequest<User[]>('/users'),
        adminRequest<Announcement[]>('/announcements'),
        adminRequest<EmailLog[]>('/email-logs'),
      ]);
      setOverview(stats); setAnalytics(usage); setUsers(people); setAnnouncements(notices); setLogs(mail);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Could not load admin data.');
    } finally { setLoading(false); }
  }, []);

  useEffect(() => { void refresh(); }, [refresh]);

  async function createAnnouncement(event: FormEvent) {
    event.preventDefault(); setBusy(true);
    try {
      await adminRequest<Announcement>('/announcements', {
        method: 'POST', body: JSON.stringify({ title, subject: subject || title, body, in_app: inApp, email }),
      });
      setTitle(''); setSubject(''); setBody('');
      toast.success('Announcement saved as a draft.');
      await refresh();
    } catch (error) { toast.error(error instanceof Error ? error.message : 'Could not save announcement.'); }
    finally { setBusy(false); }
  }

  async function publish(id: string) {
    setBusy(true);
    try {
      const result = await adminRequest<{ emails_sent: number; emails_failed: number }>(`/announcements/${id}/publish`, { method: 'POST' });
      toast.success(`Published. ${result.emails_sent} email${result.emails_sent === 1 ? '' : 's'} sent${result.emails_failed ? `; ${result.emails_failed} failed` : ''}.`);
      await refresh();
    } catch (error) { toast.error(error instanceof Error ? error.message : 'Could not publish announcement.'); }
    finally { setBusy(false); }
  }

  async function sendUsage(period: 'daily' | 'weekly' | 'monthly') {
    setBusy(true);
    try {
      const result = await adminRequest<{ users_queued: number }>(`/reports/usage?period=${period}`, { method: 'POST' });
      toast.success(`${period[0].toUpperCase()}${period.slice(1)} summary queued for ${result.users_queued} users.`);
      await refresh();
    } catch (error) { toast.error(error instanceof Error ? error.message : 'Could not send usage report.'); }
    finally { setBusy(false); }
  }

  const filteredUsers = users.filter(user => `${user.full_name} ${user.email || ''} ${user.plan}`.toLowerCase().includes(query.toLowerCase()));
  const tabs: { id: Tab; label: string; icon: typeof Activity }[] = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'announcements', label: 'Announcements', icon: Megaphone },
    { id: 'users', label: 'Users', icon: Users },
    { id: 'delivery', label: 'Email delivery', icon: Mail },
  ];

  return <AppShell>
    <main className="workspace-canvas mx-auto max-w-7xl px-5 py-8 lg:px-10 lg:py-11">
      <div className="flex flex-wrap items-end justify-between gap-5">
        <div><p className="ledger-label"><span /> ADMINISTRATION</p><h1 className="mt-3 text-4xl font-extrabold tracking-tight text-[#171717]">Control room</h1><p className="mt-2 max-w-2xl text-sm leading-6 text-[#525252]">Manage FUPISHA announcements, review accounts, and send usage summaries.</p></div>
        <Button variant="outline" onClick={() => void refresh()} disabled={loading} className="gap-2"><Activity size={16} /> Refresh data</Button>
      </div>

      <div className="mt-8 flex gap-2 overflow-x-auto border-b border-[#E5E7EB]" role="tablist" aria-label="Administration sections">
        {tabs.map(({ id, label, icon: Icon }) => <button key={id} role="tab" aria-selected={tab === id} onClick={() => setTab(id)} className={`inline-flex shrink-0 items-center gap-2 border-b-2 px-3 py-3 text-sm font-semibold transition-colors ${tab === id ? 'border-[#004AAD] text-[#004AAD]' : 'border-transparent text-[#737373] hover:text-[#171717]'}`}><Icon size={16} />{label}</button>)}
      </div>

      {loading ? <div className="flex min-h-64 items-center justify-center text-[#004AAD]" aria-live="polite"><Loader2 className="animate-spin" /><span className="sr-only">Loading administration data</span></div> : !overview ? <div className="mt-8 rounded-lg border border-amber-200 bg-amber-50 p-5 text-sm text-amber-950"><CircleAlert className="mb-2" size={18} /><p>This account could not open the administration API. Sign in with a superadmin account and retry.</p></div> : <>
        {tab === 'overview' && <>
          <section className="mt-6 grid gap-4 sm:grid-cols-3" aria-label="Platform totals">
            {[['Registered users', overview.users, Users], ['Active CVs', analytics?.totals.active_cvs ?? 0, Activity], ['Active users · 7 days', analytics?.totals.active_users_7d ?? 0, Users], ['New CVs · 30 days', analytics?.totals.new_cvs_30d ?? 0, Activity], ['Announcements', overview.announcements, Bell], ['Email records', overview.emails, Mail]].map(([label, value, Icon]: any) => <div key={label} className="rounded-lg border border-[#E5E7EB] bg-white p-5"><div className="flex items-center justify-between text-[#737373]"><span className="text-sm font-medium">{label}</span><Icon size={18} /></div><p className="mt-3 text-3xl font-bold tabular-nums text-[#171717]">{value}</p></div>)}
          </section>
          <section className="mt-6 rounded-lg border border-[#E5E7EB] bg-white p-6">
            <div className="flex flex-wrap items-start justify-between gap-4"><div><h2 className="text-lg font-bold text-[#171717]">Send a usage summary</h2><p className="mt-1 max-w-xl text-sm leading-6 text-[#525252]">Queue a one-time snapshot of each user’s active CV count. Choose the reporting period for the email subject and copy.</p></div><Activity className="text-[#004AAD]" size={20} /></div>
            <div className="mt-5 flex flex-wrap gap-2">{(['daily','weekly','monthly'] as const).map(period => <Button key={period} variant="outline" disabled={busy} onClick={() => void sendUsage(period)} className="capitalize">{busy && <Loader2 className="mr-2 animate-spin" size={15} />}{period}</Button>)}</div>
          </section>
          <section className="mt-6"><div className="mb-3 flex items-center justify-between"><h2 className="text-lg font-bold">Recent announcements</h2><Button variant="ghost" onClick={() => setTab('announcements')}>Manage announcements</Button></div><AnnouncementList items={announcements.slice(0, 5)} publish={publish} busy={busy} /></section>
        </>}
        {tab === 'announcements' && <div className="mt-6 grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(340px,.78fr)]">
          <section><div className="mb-3 flex items-center gap-2"><Megaphone size={18} className="text-[#004AAD]" /><h2 className="text-lg font-bold">Announcement history</h2></div><AnnouncementList items={announcements} publish={publish} busy={busy} /></section>
          <form onSubmit={createAnnouncement} className="space-y-5 rounded-lg border border-[#E5E7EB] bg-white p-6">
            <div><h2 className="text-lg font-bold">Write an announcement</h2><p className="mt-1 text-sm leading-6 text-[#525252]">Save a draft first. Publishing makes it visible in the app and sends email to users when selected.</p></div>
            <div className="space-y-2"><Label htmlFor="announce-title">Title</Label><Input id="announce-title" required value={title} onChange={e => setTitle(e.target.value)} placeholder="A clear announcement title" /></div>
            <div className="space-y-2"><Label htmlFor="announce-subject">Email subject</Label><Input id="announce-subject" value={subject} onChange={e => setSubject(e.target.value)} placeholder="Uses the title if left blank" /></div>
            <div className="space-y-2"><Label htmlFor="announce-body">Message</Label><Textarea id="announce-body" required rows={6} value={body} onChange={e => setBody(e.target.value)} placeholder="Write the message users need to know." /></div>
            <fieldset className="space-y-3"><legend className="text-sm font-semibold">Delivery channels</legend><label className="flex items-center gap-3 text-sm"><input type="checkbox" checked={inApp} onChange={e => setInApp(e.target.checked)} className="size-4 accent-[#004AAD]" /> In-app announcement</label><label className="flex items-center gap-3 text-sm"><input type="checkbox" checked={email} onChange={e => setEmail(e.target.checked)} className="size-4 accent-[#004AAD]" /> Email users from noreply@fupisha.link</label></fieldset>
            <Button type="submit" disabled={busy || (!inApp && !email)} className="w-full gap-2 bg-[#004AAD] text-white hover:bg-[#003A86]">{busy ? <Loader2 className="animate-spin" size={16} /> : <Plus size={16} />} Save draft</Button>
          </form>
        </div>}
        {tab === 'users' && <section className="mt-6 overflow-hidden rounded-lg border border-[#E5E7EB] bg-white"><div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#E5E7EB] p-5"><div><h2 className="text-lg font-bold">User directory</h2><p className="mt-1 text-sm text-[#737373]">{filteredUsers.length} account{filteredUsers.length === 1 ? '' : 's'}</p></div><Input value={query} onChange={e => setQuery(e.target.value)} placeholder="Search name, email, or plan" className="max-w-sm" aria-label="Search users" /></div><div className="overflow-x-auto"><table className="w-full min-w-[620px] text-left text-sm"><thead className="bg-[#F7F7F7] text-xs uppercase tracking-wide text-[#737373]"><tr><th className="px-5 py-3">User</th><th className="px-5 py-3">Plan</th><th className="px-5 py-3">Role</th></tr></thead><tbody className="divide-y divide-[#E5E7EB]">{filteredUsers.map(user => <tr key={user.id}><td className="px-5 py-4"><div className="font-semibold">{user.full_name || 'Unnamed user'}</div><div className="mt-1 text-[#737373]">{user.email || 'No email on file'}</div></td><td className="px-5 py-4 capitalize">{user.plan}</td><td className="px-5 py-4">{user.role}</td></tr>)}</tbody></table>{!filteredUsers.length && <p className="p-8 text-center text-sm text-[#737373]">No users match that search.</p>}</div></section>}
        {tab === 'delivery' && <section className="mt-6 overflow-hidden rounded-lg border border-[#E5E7EB] bg-white"><div className="border-b border-[#E5E7EB] p-5"><h2 className="text-lg font-bold">Email delivery log</h2><p className="mt-1 text-sm text-[#737373]">Announcement and usage summary send attempts.</p></div><div className="overflow-x-auto"><table className="w-full min-w-[620px] text-left text-sm"><thead className="bg-[#F7F7F7] text-xs uppercase tracking-wide text-[#737373]"><tr><th className="px-5 py-3">Recipient</th><th className="px-5 py-3">Message</th><th className="px-5 py-3">Status</th><th className="px-5 py-3">Time</th></tr></thead><tbody className="divide-y divide-[#E5E7EB]">{logs.map(log => <tr key={log.id}><td className="px-5 py-4">{log.to}</td><td className="px-5 py-4">{log.kind.replaceAll('_',' ')}</td><td className="px-5 py-4 capitalize">{log.status}</td><td className="px-5 py-4 text-[#737373]">{dateLabel(log.created_at)}</td></tr>)}</tbody></table>{!logs.length && <p className="p-8 text-center text-sm text-[#737373]">No email activity yet.</p>}</div></section>}
      </>}
    </main>
  </AppShell>;
}

function AnnouncementList({ items, publish, busy }: { items: Announcement[]; publish: (id: string) => void; busy: boolean }) {
  if (!items.length) return <div className="rounded-lg border border-dashed border-[#D4D4D4] bg-white p-8 text-center"><Megaphone className="mx-auto text-[#737373]" size={22} /><p className="mt-3 font-semibold">No announcements yet</p><p className="mt-1 text-sm text-[#737373]">Draft an update and publish it when it is ready.</p></div>;
  return <div className="divide-y divide-[#E5E7EB] rounded-lg border border-[#E5E7EB] bg-white">{items.map(item => <article key={item.id} className="p-5"><div className="flex flex-wrap items-start justify-between gap-3"><div className="min-w-0"><h3 className="font-bold">{item.title}</h3><p className="mt-1 text-sm leading-6 text-[#525252]">{item.body}</p></div><span className={`rounded-full px-2.5 py-1 text-xs font-semibold capitalize ${item.status === 'published' ? 'bg-emerald-50 text-emerald-800' : 'bg-amber-50 text-amber-900'}`}>{item.status}</span></div><div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs text-[#737373]"><span className="flex items-center gap-2">{item.in_app && <span>In app</span>}{item.email && <span>Email</span>}<span>·</span>{dateLabel(item.created_at)}</span>{item.status === 'draft' && <Button size="sm" disabled={busy} onClick={() => publish(item.id)} className="gap-2 bg-[#004AAD] text-white hover:bg-[#003A86]"><Send size={14} /> Publish</Button>}</div></article>)}</div>;
}
