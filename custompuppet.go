// mautrix-imessage - A Matrix-iMessage puppeting bridge.
// Copyright (C) 2022 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"errors"

	"maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/bridge"
	"maunium.net/go/mautrix/id"
)

var (
	ErrMismatchingMXID = errors.New("whoami result does not match custom mxid")
)

var _ bridge.DoublePuppet = (*User)(nil)

func (user *User) SwitchCustomMXID(accessToken string, mxid id.UserID) error {
	if mxid != user.MXID {
		return errors.New("mismatching mxid")
	}
	user.DoublePuppetIntent = nil
	user.AccessToken = accessToken
	err := user.StartCustomMXID(false)
	if err != nil {
		return err
	}
	return nil
}
func (user *User) CustomIntent() *appservice.IntentAPI {
	return user.DoublePuppetIntent
}

func (user *User) ClearCustomMXID() {
	user.AccessToken = ""
	user.NextBatch = ""
	user.DoublePuppetIntent = nil
}

func (user *User) StartCustomMXID(reloginOnFail bool) error {
	newIntent, newAccessToken, err := user.bridge.DoublePuppet.Setup(user.MXID, user.AccessToken, reloginOnFail)
	if err != nil {
		user.ClearCustomMXID()
		return err
	}
	if user.AccessToken != newAccessToken {
		user.AccessToken = newAccessToken
	}
	user.DoublePuppetIntent = newIntent
	return nil
}

func (user *User) tryAutomaticDoublePuppeting() {
	user.zlog.Debug().Msg("Trying to automatically enable double puppet")
	secret := user.bridge.Config.GetDoublePuppetSecret(user.MXID)
	user.zlog.Debug().Str("secret", secret).Msg("Got secret")
	if !user.bridge.Config.CanAutoDoublePuppet(user.MXID) || user.DoublePuppetIntent != nil {
		user.zlog.Debug().Msg("Automatic double puppeting not enabled")
		return
	}
	err := user.StartCustomMXID(true)
	if err != nil {
		user.zlog.Warn().Err(err).Msg("Failed to login with shared secret for double puppeting")
	} else {
		user.zlog.Info().Msg("Successfully automatically enabled double puppet")
	}
}
