/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { ProfileHeader } from '../components/profile-header'
import { useProfile } from '../hooks'
import type { UserProfile } from '../types'

const profile: UserProfile = {
  id: 7,
  username: 'alice',
  display_name: 'Alice',
  role: 1,
  group: 'default',
  quota: 0,
  used_quota: 0,
  request_count: 0,
  status: 1,
  aff_count: 0,
  aff_quota: 0,
  aff_history_quota: 0,
  created_time: 0,
}

function ProfileCard() {
  const { profile, loading, setAvatarVersion } = useProfile()
  return (
    <ProfileHeader
      profile={profile}
      loading={loading}
      onAvatarChange={setAvatarVersion}
    />
  )
}

async function renderProfile(version = '') {
  const current = { ...profile, avatar_version: version }
  useAuthStore.getState().auth.setUser(current)
  vi.spyOn(api, 'get').mockResolvedValue({
    data: { success: true, data: current },
  })
  // jsdom does not load images. Model the browser's decoded / failed image states.
  vi.stubGlobal(
    'Image',
    class {
      complete = true
      src = ''
      get naturalWidth() {
        return this.src.includes('unavailable') ? 0 : 16
      }
    }
  )
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProfileCard />
    </QueryClientProvider>
  )
  await screen.findByRole('heading', { name: 'Alice' })
}

afterEach(() => {
  useAuthStore.getState().auth.reset()
  vi.unstubAllGlobals()
})

describe('profile avatar', () => {
  it('uploads the original file and replaces the avatar URL when its version changes', async () => {
    const user = userEvent.setup()
    const put = vi.spyOn(api, 'put')
    await renderProfile()
    expect(screen.getByText('A')).toBeVisible()
    for (const version of ['first', 'next']) {
      put.mockResolvedValueOnce({
        data: { success: true, data: { avatar_version: version } },
      })
      const file = new File(['GIF89a'], 'animated.gif', { type: 'image/gif' })
      await user.upload(
        screen.getByLabelText('Upload avatar', { selector: 'input' }),
        file
      )
      const url = `/api/user/7/avatar?v=${version}`
      await waitFor(() =>
        expect(screen.getByRole('presentation')).toHaveAttribute('src', url)
      )
      expect(
        screen.getByRole('button', { name: 'Change avatar' })
      ).toBeEnabled()
      expect(useAuthStore.getState().auth.user?.avatar_version).toBe(version)
      expect(put).toHaveBeenLastCalledWith('/api/user/self/avatar', file, {
        headers: { 'Content-Type': 'image/gif' },
      })
    }
  })

  it('keeps the original fallback after an image load failure and after confirmed deletion', async () => {
    const user = userEvent.setup()
    const remove = vi.spyOn(api, 'delete').mockResolvedValue({
      data: { success: true, data: { avatar_version: '' } },
    })
    await renderProfile('unavailable')
    expect(screen.getByText('A')).toBeVisible()
    await user.click(screen.getByRole('button', { name: 'Delete avatar' }))
    expect(remove).not.toHaveBeenCalled()
    await user.click(
      within(screen.getByRole('alertdialog')).getByRole('button', {
        name: 'Delete',
      })
    )
    await screen.findByRole('button', { name: 'Upload avatar' })
    expect(remove).toHaveBeenCalledWith('/api/user/self/avatar')
    expect(useAuthStore.getState().auth.user?.avatar_version).toBe('')
    expect(screen.getByText('A')).toBeVisible()
  })

  it('rejects an oversized file before upload and retains the previous image on a server rejection', async () => {
    const error = vi.spyOn(toast, 'error')
    const put = vi.spyOn(api, 'put').mockResolvedValue({
      data: { success: false, message: '头像格式不正确' },
    })
    await renderProfile('original')
    const input = screen.getByLabelText('Upload avatar', { selector: 'input' })
    fireEvent.change(input, {
      target: {
        files: [new File([new Uint8Array(2 * 1024 * 1024 + 1)], 'big.gif')],
      },
    })
    expect(error).toHaveBeenCalledWith('Avatar must be 2 MB or smaller')
    expect(put).not.toHaveBeenCalled()
    fireEvent.change(input, {
      target: { files: [new File(['invalid'], 'invalid.gif')] },
    })
    await waitFor(() => expect(error).toHaveBeenCalledWith('头像格式不正确'))
    expect(screen.getByRole('presentation')).toHaveAttribute(
      'src',
      '/api/user/7/avatar?v=original'
    )
    expect(useAuthStore.getState().auth.user?.avatar_version).toBe('original')
  })
})
