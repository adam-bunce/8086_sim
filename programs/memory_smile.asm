bits 16
; draws a face into memory, view outputted .DATA file as 64x64 RGBA with a 256 offset from start

; fill canvas with yellow
mov bp, 64*4 ; 64 * 4 is a row, avoid overwriting instructions
mov dx, 0
y_fill_loop:
    mov cx, 0
    x_fill_loop:
        mov byte [bp + 0], 242 ; R
        mov byte [bp + 1], 255 ; G
        mov byte [bp + 2], 0   ; B
        mov byte [bp + 3], 255 ; A

        add bp, 4

        add cx, 1
        cmp cx, 64
        jnz x_fill_loop

    add dx, 1
    cmp dx, 64
    jnz y_fill_loop

; image ends at 16640
mov di, 17000

; left eye
mov word [di + 0], 16    ; length
mov word [di + 2], 16    ; width
mov word [di + 4], 5*4   ; x
mov word [di + 6], 8*256 ; y

; right eye
mov word [di + 8], 16
mov word [di + 10], 16
mov word [di + 12], 43*4
mov word [di + 14], 8*256

; mouth
mov word [di + 16], 54
mov word [di + 18], 5
mov word [di + 20], 5*4
mov word [di + 22], 48*256

mov ax, 3 ; number of rectangles in memory
sub di, 8 ; reset for inc loop
rect_draw_loop:
    sub ax, 1 ; rectangles drawn count
    add di, 8 ; adjust di to grab next rectangles values

    mov si, [di + 6] ; setup vertical start position
    mov dx, 0        ; vertical lines drawn
    y_draw_loop:
        mov cx, 0        ; horizontal pixels drawn
        mov bp, [di + 4] ; setup horizontal start position

        x_draw_loop:
            mov byte [bp + si + 0], 0   ; R
            mov byte [bp + si + 1], 0   ; G
            mov byte [bp + si + 2], 0   ; B
            mov byte [bp + si + 3], 255 ; A

            add bp, 4 ; move pixel location forward

            add cx, 1 ; increment pixels drawn
            cmp cx, [di + 0] ; check if width met
            jnz x_draw_loop

        add si, 256 ; move down a row vertical position
        add dx, 1 ; increment vertical lines drawn
        cmp dx, [di + 2] ; check if met height
        jnz y_draw_loop

    ; check if there's more rectangles to draw
    cmp ax, 0
    jnz rect_draw_loop
