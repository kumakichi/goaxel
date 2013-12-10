/*                                                                              
 * Copyright (C) 2013 Deepin, Inc.                                                 
 *               2013 Leslie Zhai <zhaixiang@linuxdeepin.com>                   
 *                                                                              
 * This program is free software: you can redistribute it and/or modify         
 * it under the terms of the GNU General Public License as published by         
 * the Free Software Foundation, either version 3 of the License, or            
 * any later version.                                                           
 *                                                                              
 * This program is distributed in the hope that it will be useful,              
 * but WITHOUT ANY WARRANTY; without even the implied warranty of               
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                
 * GNU General Public License for more details.                                 
 *                                                                              
 * You should have received a copy of the GNU General Public License            
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.        
 */

package emule

/* TODO: amule-2.3.1/src/include/protocol/ed2k/Client2Server/TCP.h */
const (
    BUFFER_SIZE         int     = 1024
    OP_EDONKEYHEADER    byte    = 0xE3
    OP_LOGINREQUEST     byte    = 0x01
    OP_SERVERMESSAGE    byte    = 0x38
    TAG_STRING          byte    = 0x02
    TAG_INTEGER         byte    = 0x03
    SPECIAL_TAG_NAME    byte    = 0x01
    SPECIAL_TAG_VERSION byte    = 0x11
    SPECIAL_TAG_PORT    byte    = 0x0F
    MET_HEADER          byte    = 0x0E
)
