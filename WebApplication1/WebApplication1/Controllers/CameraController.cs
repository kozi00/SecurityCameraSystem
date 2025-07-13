using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.SignalR;

[ApiController]
public class CameraController : ControllerBase
{
    private readonly IHubContext<CameraHub> _hub;

    public CameraController(IHubContext<CameraHub> hub)
    {
        _hub = hub;
    }

    [HttpPost("/upload")]
    public async Task<IActionResult> Upload([FromQuery] string camera)
    {
        
        using var ms = new MemoryStream();
        await Request.Body.CopyToAsync(ms);
        var bytes = ms.ToArray();
        var base64 = Convert.ToBase64String(bytes);

        Console.WriteLine($"Received frame {bytes}, base64 length: {base64.Length}");

        await _hub.Clients.All.SendAsync("ReceiveFrame", camera, base64);
        return Ok();
    }
}
