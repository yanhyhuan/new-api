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
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useUserDisplay } from '@/hooks/use-user-display'
import { getUserAvatarUrl } from '@/lib/avatar'
import type { AuthUser } from '@/stores/auth-store'

export function MobileUserSummary({ user }: { user: AuthUser }) {
  const { displayName, initials, roleLabel } = useUserDisplay(user)

  return (
    <div className='border-border flex items-center gap-2.5 border-b p-2.5 text-sm'>
      <Avatar className='size-9'>
        <AvatarImage src={getUserAvatarUrl(user)} alt='' />
        <AvatarFallback className='text-xs'>{initials}</AvatarFallback>
      </Avatar>
      <div className='flex flex-1 flex-col gap-0.5 overflow-hidden'>
        <p className='text-foreground truncate font-medium'>{displayName}</p>
        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground text-xs'>{roleLabel}</span>
          {user.group && (
            <>
              <span className='text-muted-foreground text-xs'>·</span>
              <span className='text-muted-foreground text-xs'>
                {String(user.group)}
              </span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
