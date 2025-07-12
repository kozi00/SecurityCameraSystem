using Microsoft.AspNetCore.SignalR;

public class CameraHub : Hub
{
    public async Task SendFrame(string base64Image)
    {
        await Clients.All.SendAsync("ReceiveFrame", base64Image);
    }
}
