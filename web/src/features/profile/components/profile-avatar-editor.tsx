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
import { useMutation } from '@tanstack/react-query'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { handleServerError } from '@/lib/handle-server-error'
import { createServerError } from '@/lib/server-error-message'

import { updateAvatar } from '../api'
import type { UserProfile } from '../types'

interface ProfileAvatarEditorProps {
  profile: Pick<UserProfile, 'id' | 'avatar_version'>
  onAvatarChange: (userId: number, version: string) => void
}

export function ProfileAvatarEditor(props: ProfileAvatarEditorProps) {
  const { t } = useTranslation()
  const input = useRef<HTMLInputElement>(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const mutation = useMutation({
    mutationFn: async (file: File | null) => {
      const response = await updateAvatar(file)
      if (!response.success || !response.data) {
        throw createServerError(response, t('Failed to update profile'))
      }
      return response.data.avatar_version
    },
    onSuccess: (version) => {
      props.onAvatarChange(props.profile.id, version)
      setConfirmDelete(false)
      toast.success(t('Profile updated successfully'))
    },
    onError: (error) => handleServerError(error, t('Failed to update profile')),
  })

  return (
    <div className='mt-3 space-y-1.5'>
      <div className='flex flex-wrap gap-2'>
        <Button
          type='button'
          variant='outline'
          size='sm'
          disabled={mutation.isPending}
          onClick={() => input.current?.click()}
        >
          {props.profile.avatar_version
            ? t('Change avatar')
            : t('Upload avatar')}
        </Button>
        <input
          ref={input}
          type='file'
          hidden
          aria-label={t('Upload avatar')}
          accept='.png,.apng,.jpg,.jpeg,.webp,.gif,image/png,image/apng,image/jpeg,image/webp,image/gif'
          onChange={(event) => {
            const file = event.target.files?.[0]
            event.target.value = ''
            if (!file) return
            if (file.size > 2 * 1024 * 1024) {
              toast.error(t('Avatar must be 2 MB or smaller'))
              return
            }
            mutation.mutate(file)
          }}
        />
        {props.profile.avatar_version && (
          <Button
            type='button'
            variant='ghost'
            size='sm'
            disabled={mutation.isPending}
            onClick={() => setConfirmDelete(true)}
          >
            {t('Delete avatar')}
          </Button>
        )}
      </div>
      <p className='text-muted-foreground text-xs'>
        {t('PNG, JPG, WebP or GIF, up to 2 MB. Animated images are supported.')}
      </p>
      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title={t('Delete avatar')}
        desc={t('Your initials will be shown again.')}
        destructive
        confirmText={t('Delete')}
        isLoading={mutation.isPending}
        handleConfirm={() => mutation.mutate(null)}
      />
    </div>
  )
}
