(() => {
    const dataInfo = {
        239: ["Ignition", {0: "Off", 1: "On"}],
        240: ["Movement", {0: "Off", 1: "On"}],
        21: ["GSM Signal"],
        200: ["Sleep Mode", {
            0: "No Sleep",
            1: "GPS Sleep",
            2: "Deep Sleep",
            3: "Online Sleep",
            4: "Ultra Sleep",
        }],
        69: ["GNSS Status", {
            0: "GNSS OFF",
            1: "GNSS ON with fix",
            2: "GNSS ON without fix",
            3: "GNSS sleep",
        }],
        181: ["GNSS PDOP"],
        182: ["GNSS HDOP"],
        66: ["External Voltage"],
        67: ["Battery Voltage"],
        68: ["Battery Current"],
        241: ["Active GSM Operator"],
        16: ["Total Odometer"],
        175: ["Auto Geofence", {0: "target left zone", 1: "target entered zone"}],
        252: ["Unplug", {0: "battery present", 1: "battery unplugged"}],
    }
    const prettifyData = data => {
        const pdata = { ...data }
        for (const [key, value] of Object.entries(pdata)) {
            if (key === "evt" && value in dataInfo) {
                pdata[key] = dataInfo[value][0]
                delete pdata[key]
            } else if (key in dataInfo) {
                pdata[dataInfo[key][0]] = dataInfo[key][1] ? dataInfo[key][1][value] : value
                delete pdata[key]
            }
        }
        pdata.latlng = `<a target='_blank' href='https://www.google.com/maps/place/${pdata.latlng}'>${pdata.latlng}</a>`
        return pdata
    }
    const map = L.map("map")
    const records = []
    const markers = {}
    const $day = $("#day")
    const $from = $("#from")
    const $to = $("#to")
    const $imei = $("#imei")
    const $records = $("#records")
    const MinOpacity = 0.2
    const URLParams = new URLSearchParams(window.location.hash.slice(1))

    const sleep = ms => {
        let t
        const cancel = _ => {
            clearTimeout(t)
        }
        const p = new Promise(resolve => {
            t = setTimeout(resolve, ms)
        })
        return [ p, cancel ]
    }

    const makeDate = s => {
        const d = s ? new Date(s) : new Date()
        d.setHours(0, 0, 0, 0)
        return d
    }
    const date2str = d => {
        const day = d.getDate().toString().padStart(2, "0")
        const month = (d.getMonth() + 1).toString().padStart(2, "0")
        const year = d.getFullYear()
        return `${year}-${month}-${day}`
    }

    const updateRecords = async (from, to) => {
        $imei.empty()
        records.splice(0, records.length)
        $records.html("<option disabled>Loading...</option>")
        for (const m of Object.values(markers)) {
            m.remove()
        }
        for (const latlng of Object.keys(markers)) {
            delete markers[latlng]
        }

        const res = await fetch(`/records?from=${date2str(from)}&to=${date2str(to)}`)
        records.push(...await res.json())
        $records.empty()

        let centered = false
        records.forEach((record, i) => {
            if ($imei.find(`option[value="${record.imei}"]`).length === 0) {
                $imei.append(`<option value="${record.imei}">${record.imei}</option>`)
            }
            record.created_at = new Date(record.created_at).toLocaleString()
            record.payload = JSON.parse(record.payload)
            const data = record.payload.state.reported
            const event = data.evt > 0 && data.evt in dataInfo ? ` ${dataInfo[data.evt][0]} ` : ""
            $records.append(`<option value="${record.id}">${record.created_at} >${event}> ${data.latlng}</option>`)
            if (data.latlng in markers) {
                return
            }
            const marker = L.marker(data.latlng.split(","), {opacity: MinOpacity})
            marker.addTo(map)
            marker._imei = record.imei
            marker._recordId = record.id.toString()
            marker.openRecordPopup = r => {
                marker.unbindPopup()
                const jsonData = JSON.stringify(prettifyData(r.payload.state.reported), null, 2)
                marker.bindPopup(`<pre>${r.created_at}\n${jsonData}</pre>`)
                marker.openPopup()
                URLParams.delete("record")
                URLParams.append("record", r.id)
                window.location.hash = URLParams.toString()
            }
            marker.on("click", _ => {
                marker.openRecordPopup(record)
                $records.val(marker._recordId)
                setTimeout(_ => {
                    $records.scrollTop(i * 17)
                }, 0)
            })
            marker.on("popupopen", _ => {
                marker.setOpacity(1)
            })
            marker.on("popupclose", () => {
                marker.setOpacity(MinOpacity)
                URLParams.delete("record")
                window.location.hash = URLParams.toString()
            })
            markers[data.latlng] = marker
            if (!centered) {
                map.setView(marker.getLatLng(), 14)
                centered = true
            }
        })

        const recordId = URLParams.get("record")
        if (recordId) {
            $records.val(recordId).change()
        }
    }

    const [ playAnimation, cancelAnimation ] = (_ => {
        const cancels = []
        let isCanceled = false

        const cancelAll = () => {
            isCanceled = true
            cancels.forEach(cancel => {
                cancel()
            })
            cancels.splice(0, cancels.length)
            isCanceled = false
        }

        const play = _ => {
            cancelAll()
            const speed = parseInt($("#speed").val())
            Object.values(markers).forEach(async (marker, i, markers) => {
                await (async (i) => {
                    if (isCanceled) {
                        return
                    }
                    marker.setOpacity(MinOpacity)
                    const [p, cancel] = sleep(i * speed)
                    cancels.push(cancel)
                    await p
                    marker.setOpacity(1.0)
                    markers[i-1] && markers[i-1].setOpacity(0.75)
                    markers[i-2] && markers[i-2].setOpacity(0.5)
                    markers[i-3] && markers[i-3].setOpacity(0.25)
                    markers[i-4] && markers[i-4].setOpacity(MinOpacity)
                    map.setView(marker.getLatLng())
                    $records.val(marker._recordId)
                    setTimeout(() => {
                        $records.scrollTop(i * 17)
                    }, 0)
                    if (marker === markers[-1]) {
                        cancels.splice(0, cancels.length)
                    }
                })(i)
            })
        }

        return [ play, cancelAll ]
    })()

    $day.on("change", async _ => {
        const fromVal = $day.val()
        $from.val(fromVal)
        const to = new Date(fromVal)
        to.setDate(to.getDate() + 1)
        $to.val(date2str(to)).change()
    })

    $from.add($to).on("change", async _ => {
        const from = $from.val()
        const to = $to.val()
        if (!from || !to) {
            return
        }
        await updateRecords(makeDate(from), makeDate(to))
        URLParams.delete("from")
        URLParams.append("from", from)
        URLParams.delete("to")
        URLParams.append("to", to)
        window.location.hash = URLParams.toString()
    })

    $imei.on("change", async _ => {
        const selected = $imei.val()
        Object.values(markers).forEach(marker => {
            if (selected.length > 0) {
                marker.setOpacity(0)
                if (selected.includes(marker._imei)) {
                    marker.setOpacity(MinOpacity)
                }
            } else {
                marker.setOpacity(MinOpacity)
            }
        })
    })

    $("#play").click(_ => {
        playAnimation()
    })

    $("#cancelPlay").click(_ => {
        cancelAnimation()
    })

    $records.on("change", _ => {
        const selected = $records.val()
        if (selected.length === 0) {
            return
        }
        let record
        for (const r of records) {
            if (r.id.toString() === selected[selected.length-1]) {
                record = r
                break
            }
        }
        if (!record) {
            throw new Error("WTF!? record not found!")
        }
        for (const [latlng, m] of Object.entries(markers)) {
            if (latlng === record.payload.state.reported.latlng) {
                m.openRecordPopup(record)
                const coords = m.getLatLng()
                if (coords.lat !== 0 || coords.lng !== 0) {
                    map.setView(coords)
                }
                break
            }
        }
    })

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        minZoom: 3,
        maxZoom: 19,
    }).addTo(map)

    if (window.location.hash.length > 1)  {
        $from.val(URLParams.get("from"))
        $to.val(URLParams.get("to")).change()
    } else {
        const today = makeDate()
        const tomorrow = makeDate()
        tomorrow.setDate(tomorrow.getDate() + 1)
        $from.val(date2str(today))
        $to.val(date2str(tomorrow)).change()
    }
})()